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

package plan

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"go/version"
	"slices"

	"fillmore-labs.com/errortype/internal/astutil"
	"fillmore-labs.com/errortype/internal/naming/resolve"
	"fillmore-labs.com/errortype/internal/naming/rules"
)

// Candidate is the complete per-declaration output of [Plan].
type Candidate struct {
	// Obj is the declared object whose name violates the error naming convention.
	Obj types.Object

	// NewName is the suggested replacement name.
	NewName string

	// Uses lists the positions to rewrite when applying the rename. Embedding
	// sites and field uses that must keep the old field name are excluded.
	Uses []token.Pos

	// DocComment holds the position of the comment name that should be
	// renamed and also a position to insert the deprecation to, when necessary.
	DocComment astutil.DocCommentMetadata

	// Fixable reports whether the rename is safe to apply as a suggested fix,
	// false means the diagnostic is emitted without a fix.
	Fixable bool

	// Deprecate signals that a deprecated alias should be added.
	Deprecate bool
}

// Plan constructs a list of rename candidates for the violations.
//
// It builds the context needed to rename each violation: it maps use sites,
// applies generated-file handling, and resolves a conflict-free name, reusing
// the doc-comment metadata already recorded by [rules.Check].
// It then returns a plan to be executed by the reporting stage.
//
// The new name starts from the [rules.Suggest] proposal. It conflicts when the
// name is already declared or claimed by another planned rename in the same
// scope, when it is a file-block name such as an import (for package-level
// renames), when an inner declaration would shadow a use site, or when the new
// declaration would capture a reference that resolves to an outer scope. On
// conflict, numbered variants ("errX2" through "errX9", or "x2Error" through
// "x9Error" for types) are tried instead.
//
// A type embedded by name keeps its field name behind a deprecation alias.
func Plan(info *types.Info, resolver *resolve.Resolver, fileMap astutil.FileMap, violations []rules.Violation) (renames []Candidate, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = toError(r)
		}
	}()

	// Classify uses and types embedded by name in a struct field.
	usesByObj, embeddings := mapUsesByObject(info, violations)

	renames = make([]Candidate, 0, len(violations))

	builder := candidateBuilder{
		info:    info,
		fileMap: fileMap,
	}

	for _, v := range violations {
		obj := v.Obj
		uses, embedded := usesByObj[obj], embeddings.isEmbedded(obj)
		rn := builder.buildCandidate(v, uses, embedded)

		// Fill [Candidate.NewName] with a conflict-free suggestion where possible.
		rn.NewName, rn.Fixable = resolver.Name(obj, rn.Fixable, rn.Uses)

		renames = append(renames, rn)
	}

	return renames, nil
}

func toError(r any) error {
	if e, ok := r.(error); ok {
		return e
	}

	return fmt.Errorf("%#v", r)
}

type candidateBuilder struct {
	info    *types.Info
	fileMap astutil.FileMap
}

// buildCandidate handles the context-gathering logic for a single violation.
// It returns a populated [Candidate].
func (b candidateBuilder) buildCandidate(v rules.Violation, uses []token.Pos, embedded bool) Candidate {
	obj := v.Obj

	// Insert a deprecation alias for embedded or exported declarations.
	fixable, deprecate := true, embedded || obj.Exported()

	file := b.fileMap.File(obj.Pos())
	if b.fileMap.IsGenerated(file) {
		fixable = false
	}

	// Uses in generated files are not modified but aliased.
	if fixable && !deprecate && b.fileMap.HasGenerated(uses) {
		uses = slices.Clone(uses)
		uses = slices.DeleteFunc(uses, b.fileMap.InGenerated)

		deprecate = true
	}

	// Are generic type aliases required but not supported?
	if fixable && deprecate && v.DocComment.TypeParams != nil && !b.genericTypeAliasSupport(file) {
		fixable = false
	}

	return Candidate{
		Obj:        obj,
		Uses:       uses,
		DocComment: v.DocComment,
		Fixable:    fixable,
		Deprecate:  deprecate,
	}
}

// genericTypeAliasSupport reports whether a deprecation with a generic type alias
// would affect a file targeting a Go version with a least 1.24,
// where support was introduced (https://go.dev/blog/alias-names).
func (b candidateBuilder) genericTypeAliasSupport(astFile *ast.File) bool {
	goversion := b.info.FileVersions[astFile]

	return goversion == "" || version.Compare("go1.24", goversion) <= 0
}
