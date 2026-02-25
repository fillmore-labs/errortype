// Copyright 2026 Oliver Eikemeier. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package analyze

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"go/version"
	"slices"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// assertionQuery holds the components of a matched error-query helper of the form:
//
//	func (err error) bool {
//		_, ok := err.(T)
//		return ok
//	}
type assertionQuery struct {
	param        *types.Var          // the error parameter being asserted
	okVar        *types.Var          // the boolean variable bound by the assertion and returned
	typeAssert   *ast.TypeAssertExpr // the err.(T) assertion
	asgn         *ast.AssignStmt     // the "_, ok := err.(T)" statement
	retStmt      *ast.ReturnStmt     // the "return ok" statement
	result       *ast.Field          // The single result
	assertedType types.Type          // type of T in the err.(T) assertion
}

// matchAssertionQuery reports whether f is an error-query helper and, if so, returns
// its components.
func (p Pass) matchAssertionQuery(ftype *ast.FuncType, body *ast.BlockStmt) (assertionQuery, bool) {
	var q assertionQuery

	ok := q.matchBody(body) &&
		q.matchParameters(p.TypesInfo, ftype) &&
		q.matchAssertion(p.TypesInfo)

	return q, ok
}

func (q *assertionQuery) matchParameters(info *types.Info, ftype *ast.FuncType) bool {
	param := singleParam(info, ftype.Params)
	if param == nil {
		return false // not a function with a single named parameter
	}

	q.param = param

	result := singleField(ftype.Results)
	if result == nil {
		return false // not a single result
	}

	q.result = result

	return true
}

func (q *assertionQuery) matchBody(body *ast.BlockStmt) bool {
	if body == nil || len(body.List) != 2 {
		return false
	}

	asgn, ok := body.List[0].(*ast.AssignStmt)
	if !ok || len(asgn.Lhs) != 2 || len(asgn.Rhs) != 1 {
		return false
	}

	q.asgn = asgn

	typeAssert, ok := ast.Unparen(asgn.Rhs[0]).(*ast.TypeAssertExpr)
	if !ok {
		return false
	}

	q.typeAssert = typeAssert

	retStmt, ok := body.List[1].(*ast.ReturnStmt)
	if !ok {
		return false
	}

	q.retStmt = retStmt

	return true
}

func (q *assertionQuery) matchAssertion(info *types.Info) bool {
	if !typeutil.IsInterfaceWithError(q.param.Type()) {
		return false
	}

	if rtyp, ok := info.Types[q.result.Type]; !ok || rtyp.Type != types.Typ[types.Bool] {
		return false
	}

	if !checkAssert(info, q.param, q.typeAssert.X) {
		return false
	}

	okVar := okResultVar(info, q.asgn)
	if okVar == nil {
		return false
	}

	q.okVar = okVar

	if retVar := returnedVar(info, q.retStmt.Results, q.result); retVar != okVar {
		return false
	}

	tv, ok := info.Types[q.typeAssert.Type]
	if !ok || !tv.IsType() {
		return false
	}

	q.assertedType = tv.Type

	return true
}

// singleParam checks if a field list contains exactly one field, which itself has one name.
func singleParam(info *types.Info, fields *ast.FieldList) *types.Var {
	if fields == nil || len(fields.List) != 1 {
		return nil
	}

	field := fields.List[0]
	if len(field.Names) != 1 {
		return nil
	}

	id := field.Names[0]
	if id == nil || id.Name == "_" {
		return nil
	}

	param, _ := info.Defs[id].(*types.Var)

	return param
}

// singleField returns the sole result field of results (optionally named), or nil when
// results does not hold exactly one field with at most one name.
func singleField(fields *ast.FieldList) *ast.Field {
	if fields == nil || len(fields.List) != 1 {
		return nil
	}

	field := fields.List[0]
	if len(field.Names) > 1 {
		return nil
	}

	return field
}

// checkAssert reports whether the asserted expression resolves to the function parameter.
func checkAssert(info *types.Info, param *types.Var, asserted ast.Expr) bool {
	assertedID, ok := ast.Unparen(asserted).(*ast.Ident)
	if !ok {
		return false
	}

	obj, ok := info.Uses[assertedID]

	return ok && obj == param
}

// okResultVar resolves the variable bound to the boolean result of the comma-ok
// assignment asgn, or nil when the second operand is not a plain identifier.
func okResultVar(info *types.Info, asgn *ast.AssignStmt) *types.Var {
	okID, ok := ast.Unparen(asgn.Lhs[1]).(*ast.Ident)
	if !ok {
		return nil
	}

	switch asgn.Tok {
	case token.DEFINE:
		okVar, _ := info.Defs[okID].(*types.Var)

		return okVar

	case token.ASSIGN:
		okVar, _ := info.Uses[okID].(*types.Var)

		return okVar

	default:
		return nil
	}
}

// legacyFix builds the text edits that replace q's legacy type assertion with [errors.AsType]
// or [errors.As]. It returns nil when no safe rewrite is available, e.g. when the errors
// qualifier would collide with an in-scope identifier.
func (q assertionQuery) legacyFix(info *types.Info, fset *token.FileSet, file *ast.File) []analysis.TextEdit {
	qual, edits, ok := resolveErrorsImport(info, file)
	if !ok {
		return nil
	}

	useAsType := hasGo126(info, file) && typeutil.HasErrorMethod(q.assertedType)

	asFun := "AsType"
	if !useAsType {
		asFun = "As"
	}

	paramName, okName := q.param.Name(), q.okVar.Name()

	switch qual {
	case paramName, okName:
		return nil // the errors qualifier is shadowed by the parameter or returned variable

	default:
		if asFun == paramName || asFun == okName {
			return nil // the dot-imported function is shadowed by the parameter or returned variable
		}
	}

	call, typ := qualify(qual, asFun), exprToString(fset, q.typeAssert.Type)

	// Go1.26+: errors.AsType[T](err) keeps the value and needs no target variable.
	if useAsType {
		asType := fmt.Appendf(nil, "%s[%s](%s)", call, typ, paramName)

		return append(edits, analysis.TextEdit{
			Pos:     q.asgn.Rhs[0].Pos(),
			End:     q.asgn.End(),
			NewText: asType,
		})
	}

	var newText []byte

	// Pre-1.26: errors.As(err, <target>). Reuse the variable that captured the asserted
	// value to preserve the write; otherwise discard the value with new(T).
	switch q.asgn.Tok {
	case token.DEFINE:
		// ok is a fresh local, so collapse to a single return
		newText = fmt.Appendf(nil, "return %s(%s, new(%s))", call, paramName, typ)

	case token.ASSIGN:
		target := targetExpr(info, q.asgn, q.assertedType, typ)
		if target == "" {
			return nil
		}

		newText = fmt.Appendf(nil, "%s = %s(%s, %s)\n\treturn %s", okName, call, paramName, target, okName)

	default:
		panic(fmt.Sprintf("unexpected assignment %s", q.asgn.Tok))
	}

	return append(edits, analysis.TextEdit{
		Pos:     q.asgn.Pos(),
		End:     q.retStmt.End(),
		NewText: newText,
	})
}

// targetExpr returns the second argument to errors.As from assignment asgn: a pointer
// to the variable that captured the asserted value (preserving the write in a success case,
// but leaving it untouched when not matched, an observable difference),
// or new(typ) when the value was discarded with "_".
//
// It returns "" when the destination is an expression that cannot be reused as a target.
func targetExpr(info *types.Info, asgn *ast.AssignStmt, assertedType types.Type, typ string) string {
	target, ok := ast.Unparen(asgn.Lhs[0]).(*ast.Ident)
	if !ok {
		return "" // the asserted value is written to an expression we are not sure how to reuse.
	}

	if target.Name == "_" {
		return fmt.Sprintf("new(%s)", typ)
	}

	// This protects against writing to a global with the wrong type.
	if v, ok := info.Uses[target]; !ok || !types.Identical(v.Type(), assertedType) {
		return ""
	}

	return "&" + target.Name
}

// qualify renders a reference to name in the errors package: "errors.As" for a normal
// import, or just "As" for a dot-import (qual == ".").
func qualify(qual, name string) string {
	if qual == "." {
		return name
	}

	return qual + "." + name
}

// resolveErrorsImport returns the qualifier to use for the errors package and any edit
// needed to import it. When the package is not already imported it adds an import, using the
// name "errors" or, when that is taken anywhere in scope, the alias errorsAlias. ok is false
// only when both names are taken and the package cannot be referenced.
func resolveErrorsImport(info *types.Info, file *ast.File) (qual string, edits []analysis.TextEdit, ok bool) {
	switch spec := errorsImportSpec(file); {
	case spec == nil:
		// Not imported under a usable name; add it, falling back to errorsAlias when the
		// "errors" name is already taken.
		for _, packageName := range []string{errorsImport, errorsAlias} {
			if _, obj := info.Scopes[file].LookupParent(packageName, token.NoPos); obj == nil {
				alias := ""
				if packageName != errorsImport {
					alias = packageName
				}

				return packageName, []analysis.TextEdit{addImportEdit(file, alias, errorsImportPath)}, true
			}
		}

		return "", nil, false

	case spec.Name == nil:
		return errorsImport, nil, true

	default:
		// Handles also the special case of spec.Name.Name == "."
		return spec.Name.Name, nil, true
	}
}

// errorsImportSpec returns the import spec for the "errors" package in file,
// or nil when the package is absent or imported only as a blank import.
func errorsImportSpec(file *ast.File) *ast.ImportSpec {
	i := slices.IndexFunc(file.Imports, isErrorImport)
	if i < 0 {
		return nil
	}

	return file.Imports[i]
}

// isErrorImport reports whether spec imports the standard "errors" package under a usable
// (non-blank) name.
func isErrorImport(spec *ast.ImportSpec) bool {
	return spec.Path.Value == errorsImportPath && (spec.Name == nil || spec.Name.Name != "_")
}

const (
	errorsImport     = "errors"
	errorsImportPath = `"` + errorsImport + `"`

	// errorsAlias names the standard errors import when "errors" is already taken, e.g. by a
	// github.com/pkg/errors import.
	errorsAlias = "goerrors"
)

// addImportEdit returns a TextEdit that imports importPath, under the local name packageName
// when it is non-empty (e.g. `packageName "importPath"`). The import is inserted into file's
// grouped import block, or as a new import declaration after the package clause when no
// grouped block exists.
func addImportEdit(file *ast.File, packageName, importPath string) analysis.TextEdit {
	var decl *ast.GenDecl
	if len(file.Decls) > 0 {
		decl, _ = file.Decls[0].(*ast.GenDecl)
	}

	importLine, pos := findImport(decl, file)

	var newText bytes.Buffer

	newText.WriteString(importLine)

	if packageName != "" {
		newText.WriteString(packageName)
		newText.WriteByte(' ')
	}

	newText.WriteString(importPath)

	return analysis.TextEdit{
		Pos:     pos,
		End:     pos,
		NewText: newText.Bytes(),
	}
}

func findImport(decl *ast.GenDecl, file *ast.File) (string, token.Pos) {
	if decl != nil && decl.Tok == token.IMPORT && decl.Lparen.IsValid() {
		// If the file already has imports, prefix to the first import declaration.
		// Create a TextEdit that inserts the new import just after the opening parenthesis.
		return "\n\t", decl.Lparen + 1
	}

	// Fallback: No grouped import block. Insert right after the package declaration.
	return "\n\nimport ", file.Name.End()
}

// hasGo126 reports whether file is compiled with language version go1.26 or later,
// which is required for the generic errors.AsType helper.
func hasGo126(info *types.Info, file *ast.File) bool {
	goversion, ok := info.FileVersions[file]
	return ok && version.Compare(goversion, "go1.26") >= 0
}

// returnedVar returns the *[types.Var] returned by retStmt, handling bare named-return and explicit return.
func returnedVar(info *types.Info, returned []ast.Expr, fResult *ast.Field) *types.Var {
	if len(returned) == 0 { // Bare return
		if len(fResult.Names) == 0 {
			return nil // No named return values
		}

		retID := fResult.Names[0]
		retVar, _ := info.Defs[retID].(*types.Var)

		return retVar
	}

	retID, ok := ast.Unparen(returned[0]).(*ast.Ident)
	if !ok {
		return nil // some other expression
	}

	retVar, _ := info.Uses[retID].(*types.Var)

	return retVar
}
