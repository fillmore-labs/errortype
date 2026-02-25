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

package astutil

import (
	"go/ast"
	"go/token"
	"go/types"
)

// DocCommentMetadata describes the doc comment of a declaration that is about to be
// renamed: where to insert a deprecation notice, and where its name is
// mentioned, so they can be rewritten alongside the declaration.
type DocCommentMetadata struct {
	// TypeParams is the type-parameter list of the enclosing [ast.TypeSpec], or nil.
	TypeParams *ast.FieldList

	// Mentions holds the positions of every standalone mention of the declared
	// name inside the doc comment, so they can be renamed too, or is empty.
	Mentions []token.Pos

	// DeprecationPos is the insertion point for a "Deprecated:" alias
	// to preserve the old name for not renamed usage sites, or [token.NoPos].
	DeprecationPos token.Pos

	// MultiDeclaration is true when the declaration belongs to a parenthesized
	// "( … )" group, which determines the indentation and placement of an
	// inserted deprecation alias.
	MultiDeclaration bool
}

// CheckValueSpecComments inspects the enclosing declaration's doc comments.
// It returns:
//   - (info, true) with the [DocCommentMetadata] for the declaration.
//   - (DocCommentMetadata{}, false) when the diagnostic should be suppressed
//     because the declaration is already deprecated or carries a matching nolint
//     directive.
func CheckValueSpecComments(info *types.Info, spec *ast.ValueSpec, decl *ast.GenDecl, id *ast.Ident, linter string) (DocCommentMetadata, bool) {
	specDoc, specComment := spec.Doc, spec.Comment

	mentions, ok := checkMentions(specDoc, specComment, decl, id.Name, linter)
	if !ok {
		return DocCommentMetadata{}, false
	}

	var deprecationPos token.Pos

	multiDeclaration := decl.Lparen.IsValid()
	if multiDeclaration && decl.Tok == token.CONST {
		// For const declarations, inserting values will change iota order.
		multiDeclaration = !hasIota(info, decl)
	}

	switch {
	case multiDeclaration:
		deprecationPos = preComments(specDoc, spec.Pos())
	default:
		deprecationPos = preComments(decl.Doc, decl.Pos())
	}

	return DocCommentMetadata{
		DeprecationPos:   deprecationPos,
		Mentions:         mentions,
		MultiDeclaration: multiDeclaration,
	}, true
}

// CheckTypeSpecComments inspects the enclosing declaration's doc comments.
// It returns:
//   - (info, true) with the [DocCommentMetadata] for the declaration.
//   - (DocCommentMetadata{}, false) when the diagnostic should be suppressed
//     because the declaration is already deprecated or carries a matching nolint
//     directive.
func CheckTypeSpecComments(spec *ast.TypeSpec, decl *ast.GenDecl, linter string) (DocCommentMetadata, bool) {
	id := spec.Name
	specDoc, specComment := spec.Doc, spec.Comment

	mentions, ok := checkMentions(specDoc, specComment, decl, id.Name, linter)
	if !ok {
		return DocCommentMetadata{}, false
	}

	var deprecationPos token.Pos

	multiDeclaration := decl.Lparen.IsValid()
	switch {
	case multiDeclaration:
		deprecationPos = preComments(specDoc, spec.Pos())
	default:
		deprecationPos = preComments(decl.Doc, decl.Pos())
	}

	return DocCommentMetadata{
		DeprecationPos:   deprecationPos,
		Mentions:         mentions,
		MultiDeclaration: multiDeclaration,
	}, true
}

func checkMentions(specDoc, specComment *ast.CommentGroup, decl *ast.GenDecl, name, linter string) ([]token.Pos, bool) {
	if HasNoLint(specComment, linter) {
		return nil, false // nolint, don't process further
	}

	var mentions []token.Pos

	if specDoc != nil {
		positions, suppressed := AnalyzeComments(specDoc, name, linter)
		if suppressed {
			return nil, false // deprecated or nolint, don't process further
		}

		// spec.Doc is always the comment for this identifier, so the name
		// prefix (if any) is safe to rename.
		mentions = positions
	}

	switch {
	case decl.Doc == nil:

	case specDoc == nil:
		// A group comment is specific to a single declaration only when the
		// declaration has no individual comment.
		positions, suppressed := AnalyzeComments(decl.Doc, name, linter)
		if suppressed {
			return nil, false // deprecated or nolint, don't process further
		}

		mentions = positions

	default:
		if _, suppressed := AnalyzeComments(decl.Doc, "", linter); suppressed {
			return nil, false // deprecated or nolint, don't process further
		}
	}

	return mentions, true
}

func preComments(specDoc *ast.CommentGroup, fallbackPos token.Pos) token.Pos {
	if specDoc != nil {
		return specDoc.Pos()
	}

	return fallbackPos
}

var universeIota = types.Universe.Lookup("iota").(*types.Const)

func hasIota(info *types.Info, decl *ast.GenDecl) bool {
	for n := range ast.Preorder(decl) {
		id, ok := n.(*ast.Ident)
		if !ok {
			continue
		}

		if use, ok := info.Uses[id]; ok && use == universeIota {
			return true
		}
	}

	return false
}
