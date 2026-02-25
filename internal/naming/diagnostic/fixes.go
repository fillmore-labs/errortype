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

package diagnostic

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/astutil"
)

// createFixes generates suggested fixes for renaming an object and updating its references and doc comments.
// It handles declaration updates, deprecation comments, and use-site renames.
func createFixes(fset *token.FileSet, obj types.Object, newName string, uses []token.Pos, declarationComment astutil.DocCommentMetadata, deprecate bool) []analysis.SuggestedFix {
	// One edit for the declaration, plus the doc-comment mentions and use sites.
	numEdits := 1 + len(declarationComment.Mentions) + len(uses)

	if deprecate {
		numEdits++
	}

	edits := make([]analysis.TextEdit, 0, numEdits)

	if deprecate {
		// Insert deprecation
		if newText := deprecationText(fset, obj, declarationComment.TypeParams, newName, declarationComment.MultiDeclaration); newText != nil {
			edits = append(edits, deprecationInsert(declarationComment.DeprecationPos, newText))
		}
	}

	nameLen := len(obj.Name())
	newNameText := []byte(newName)

	// Rename references in the doc comment
	for _, pos := range declarationComment.Mentions {
		edits = append(edits, renameEdit(pos, nameLen, newNameText))
	}

	// Rename declaration
	edits = append(edits, renameEdit(obj.Pos(), nameLen, newNameText))

	// Rename uses
	for _, pos := range uses {
		edits = append(edits, renameEdit(pos, nameLen, newNameText))
	}

	return []analysis.SuggestedFix{{
		Message:   fmt.Sprintf("Rename %q to %q", obj.Name(), newName),
		TextEdits: edits,
	}}
}

func deprecationText(fset *token.FileSet, obj types.Object, typeParams *ast.FieldList, newName string, multiDeclaration bool) []byte {
	var (
		decl   token.Token
		inline bool
	)

	switch obj.(type) {
	case *types.Const:
		decl = token.CONST
		inline = true

	case *types.Var:
		// Note: For variables, this generates `var OldVar = NewVar` which
		// copies the value. Variables cannot be truly aliased at the language level,
		// so modifications to NewVar or address comparisons will behave differently.
		decl = token.VAR

	case *types.TypeName:
		decl = token.TYPE
		inline = true

	default:
		return nil
	}

	indent := ""
	if multiDeclaration {
		indent = "\t"
	}

	var buf bytes.Buffer
	_, _ = fmt.Fprintf(&buf, "%s// Deprecated: Use [%s] instead.\n", indent, newName) // ignore error

	if inline {
		if !multiDeclaration {
			_, _ = buf.WriteString("\t//\n") // ignore error
		}
		_, _ = fmt.Fprintf(&buf, "%s//go:fix inline\n", indent) // ignore error
	}

	if multiDeclaration {
		_, _ = fmt.Fprintf(&buf, "\t%s", obj.Name()) // ignore error
	} else {
		_, _ = fmt.Fprintf(&buf, "%s %s", decl, obj.Name()) // ignore error
	}

	writeTypeParamList(&buf, fset, typeParams)
	_, _ = fmt.Fprintf(&buf, " = %s", newName) // ignore error
	writeTypeArgList(&buf, typeParams)

	_ = buf.WriteByte('\n') // ignore error
	if !multiDeclaration {
		_ = buf.WriteByte('\n') // ignore error
	}

	return buf.Bytes()
}

var rawfmt = &printer.Config{Mode: printer.RawFormat}

// writeTypeParamList writes params with constraints, e.g. "[T any]".
func writeTypeParamList(buf *bytes.Buffer, fset *token.FileSet, params *ast.FieldList) {
	if params == nil {
		return
	}

	_ = buf.WriteByte('[') // ignore error

	first := true

	for _, field := range params.List {
		for _, n := range field.Names {
			if !first {
				_, _ = buf.WriteString(", ") // ignore error
			}

			_, _ = buf.WriteString(n.Name) // ignore error

			first = false
		}

		_ = buf.WriteByte(' ')                   // ignore error
		_ = rawfmt.Fprint(buf, fset, field.Type) // ignore error
	}

	_ = buf.WriteByte(']') // ignore error
}

// writeTypeArgList writes params as type arguments, e.g. "[T]".
func writeTypeArgList(buf *bytes.Buffer, params *ast.FieldList) {
	if params == nil {
		return
	}

	_ = buf.WriteByte('[') // ignore error

	first := true

	for _, field := range params.List {
		for _, n := range field.Names {
			if !first {
				_, _ = buf.WriteString(", ") // ignore error
			}

			_, _ = buf.WriteString(n.Name) // ignore error

			first = false
		}
	}

	_ = buf.WriteByte(']') // ignore error
}

func deprecationInsert(pos token.Pos, newText []byte) analysis.TextEdit {
	return analysis.TextEdit{
		Pos:     pos,
		End:     pos,
		NewText: newText,
	}
}

func renameEdit(pos token.Pos, nameLen int, newName []byte) analysis.TextEdit {
	return analysis.TextEdit{
		Pos:     pos,
		End:     token.Pos(int(pos) + nameLen),
		NewText: newName,
	}
}
