// Copyright 2025 Oliver Eikemeier. All Rights Reserved.
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

package report

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"go/version"
	"strings"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/qualname"
)

// Base provides the basic fields for reporting diagnostics.
// It holds the analysis pass and the AST node where the diagnostic should be reported.
type Base struct {
	*analysis.Pass
	Expr ast.Expr
}

// UndeterminedUsage reports a diagnostic for an error type with undetermined usage.
func (b Base) UndeterminedUsage(tn *types.TypeName, ptr bool) {
	fullName := qualname.NewType(tn).String()

	codeSuffix := ""
	if ptr {
		codeSuffix = "+"
	}

	switch tnPkg := tn.Pkg(); {
	case tnPkg == b.Pkg:
		firstForm, secondForm, note := assertForms(tn, "", ptr)

		msg := fmt.Sprintf(
			`Undetermined usage for error type %q. Add a package-level "var _ error = %s" or "var _ error = %s"%s. (et:emb%s)`,
			tn.Name(), firstForm, secondForm, note, codeSuffix,
		)

		var related []analysis.RelatedInformation
		if tn.Pos().IsValid() {
			related = []analysis.RelatedInformation{{
				Pos:     tn.Pos(),
				Message: fmt.Sprintf(`Declare intent for %q here: "var _ error = %s" or "var _ error = %s"%s`, tn.Name(), firstForm, secondForm, note),
			}}
		}

		b.Report(analysis.Diagnostic{Pos: b.Expr.Pos(), End: b.Expr.End(), Message: msg, Related: related})

	case inSameModule(b.Module, tnPkg):
		firstForm, secondForm, note := assertForms(tn, "", ptr)

		b.ReportRangef(b.Expr,
			`Undetermined usage for error type %q. Add a package-level "var _ error = %s" or "var _ error = %s"%s in package %q. (et:emb%s)`,
			fullName, firstForm, secondForm, note, tnPkg.Path(), codeSuffix)

	case tnPkg != nil && tn.Exported():
		firstForm, secondForm, note := assertForms(tn, tnPkg.Name()+".", ptr)

		b.ReportRangef(b.Expr,
			`Undetermined usage for error type %q. Add a package-level "var _ error = %s" or "var _ error = %s"%s. (et:emb%s)`,
			fullName, firstForm, secondForm, note, codeSuffix)

	default:
		b.ReportRangef(b.Expr,
			`Undetermined usage for error type %q. Add it under "pointer", "value", or "suppress" in overrides. (et:emb%s)`,
			fullName, codeSuffix)
	}
}

func assertForms(tn *types.TypeName, pkgPrefix string, ptr bool) (firstForm, secondForm, note string) {
	name, unresolved := definingTypeName(tn)
	name = pkgPrefix + name
	ptrForm := pointerZeroFormNamed(name)
	valForm := zeroValueFormNamed(name, tn.Type())

	if ptr {
		return ptrForm, valForm, unresolvedNote(unresolved)
	}

	return valForm, ptrForm, unresolvedNote(unresolved)
}

func unresolvedNote(unresolved []string) string {
	switch len(unresolved) {
	case 0:
		return ""

	case 1:
		return " (with a concrete type for " + unresolved[0] + ")"

	default:
		return " (with concrete types for " + strings.Join(unresolved, ", ") + ")"
	}
}

func inSameModule(mod *analysis.Module, pkg *types.Package) bool {
	if mod == nil || mod.Path == "" || pkg == nil {
		return false
	}

	pkgPath := pkg.Path()

	return pkgPath == mod.Path || strings.HasPrefix(pkgPath, mod.Path+"/")
}

var _sampleTypes = [...]struct {
	typ  types.Type
	name string
}{
	{types.Typ[types.Int], "int"},
	{types.Typ[types.String], "string"},
	{types.Universe.Lookup("error").Type(), "error"},
}

func definingTypeName(tn *types.TypeName) (name string, unresolved []string) {
	baseName := tn.Name()

	tp, ok := tn.Type().(interface{ TypeParams() *types.TypeParamList })
	if !ok {
		return baseName, nil
	}

	tparams := tp.TypeParams()
	if tparams.Len() == 0 {
		return baseName, nil
	}

	args := make([]string, tparams.Len())
	for i := range tparams.Len() {
		param := tparams.At(i)
		if sample, ok := sampleTypeArg(param.Constraint()); ok {
			args[i] = sample
			continue
		}

		paramName := param.Obj().Name()
		if paramName == "" || paramName == "_" {
			paramName = "T"
		}

		args[i] = paramName
		unresolved = append(unresolved, paramName)
	}

	return baseName + "[" + strings.Join(args, ", ") + "]", unresolved
}

func sampleTypeArg(constraint types.Type) (string, bool) {
	iface, ok := constraint.Underlying().(*types.Interface)
	if !ok {
		return "", false
	}

	for _, sample := range _sampleTypes {
		if types.Satisfies(sample.typ, iface) {
			return sample.name, true
		}
	}

	return "", false
}

func (b Base) relativeNameOf(tn *types.TypeName) string {
	return types.TypeString(tn.Type(), types.RelativeTo(b.Pkg))
}

func (b Base) shortNameOf(tn *types.TypeName) string {
	current := b.Pkg

	return types.TypeString(tn.Type(), func(pkg *types.Package) string {
		if pkg == current {
			return ""
		}

		return pkg.Name()
	})
}

// valueForm returns how a value of type tn is written in a suggestion:
// "T{...}" for types with composite literals, "T(...)" otherwise.
func (b Base) valueForm(tn *types.TypeName) string {
	if hasCompositeLiteral(tn.Type()) {
		return b.shortNameOf(tn) + "{...}"
	}

	return b.shortNameOf(tn) + "(...)"
}

// pointerForm returns how a pointer to a new value of type tn is written in a suggestion:
// "&T{...}" for types with composite literals, otherwise "new(T(...))" in files
// at Go 1.26 or later (new with an expression operand) and "new(T)" before.
func (b Base) pointerForm(tn *types.TypeName) string {
	switch name := b.shortNameOf(tn); {
	case hasCompositeLiteral(tn.Type()):
		return "&" + name + "{...}"

	case b.goVersionAtLeast("go1.26"):
		return "new(" + name + "(...))"

	default:
		return "new(" + name + ")"
	}
}

// zeroValueForm returns the zero value of type tn, as used in a compile-time
// assertion: "T{}", "T(0)", "T(\"\")", "T(false)" or "T(nil)".
func (b Base) zeroValueForm(tn *types.TypeName) string {
	return zeroValueFormNamed(b.shortNameOf(tn), tn.Type())
}

func zeroValueFormNamed(name string, typ types.Type) string {
	switch u := typ.Underlying().(type) {
	case *types.Struct, *types.Array, *types.Slice, *types.Map:
		return name + "{}"

	case *types.Basic:
		switch info := u.Info(); {
		case info&types.IsNumeric != 0:
			return name + "(0)"

		case info&types.IsString != 0:
			return name + `("")`

		case info&types.IsBoolean != 0:
			return name + "(false)"
		}
	}

	return name + "(nil)" // pointers, functions, channels and unsafe.Pointer
}

// pointerZeroForm returns the nil pointer to type tn, as used in a compile-time assertion: "(*T)(nil)".
func (b Base) pointerZeroForm(tn *types.TypeName) string {
	return pointerZeroFormNamed(b.shortNameOf(tn))
}

func pointerZeroFormNamed(name string) string {
	return "(*" + name + ")(nil)"
}

// goVersionAtLeast reports whether the file containing the diagnostic is compiled
// with language version goversion or later.
func (b Base) goVersionAtLeast(goversion string) bool {
	pos := b.Expr.Pos()
	for _, file := range b.Files {
		if file.FileStart <= pos && pos <= file.FileEnd {
			fileVersion, ok := b.TypesInfo.FileVersions[file]

			return ok && version.Compare(fileVersion, goversion) >= 0
		}
	}

	return false
}

// hasCompositeLiteral reports whether values of typ can be written as composite literals.
func hasCompositeLiteral(typ types.Type) bool {
	switch typ.Underlying().(type) {
	case *types.Struct, *types.Array, *types.Slice, *types.Map:
		return true

	default:
		return false
	}
}

var rawfmt = &printer.Config{Mode: printer.RawFormat}

func exprToString(fset *token.FileSet, expr ast.Expr) string {
	if sb := (strings.Builder{}); rawfmt.Fprint(&sb, fset, expr) == nil {
		return sb.String()
	}

	return "<invalid>"
}
