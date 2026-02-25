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

package rules

import (
	"cmp"
	"go/ast"
	"go/types"
	"slices"

	gotypeutil "golang.org/x/tools/go/types/typeutil"

	"fillmore-labs.com/errortype/internal/astutil"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// New returns a new [Checker] instance for analyzing error-related declarations.
func New(info *types.Info, linter string) *Checker {
	return &Checker{info: info, linter: linter}
}

// Checker analyzes Go code for error-typed declarations that violate the [naming conventions].
//
// [naming conventions]: https://go.dev/wiki/Errors#naming
type Checker struct {
	info       *types.Info
	linter     string
	cache      gotypeutil.Map
	violations []Violation
}

// Violation represents a found issue in the code: an error-typed declaration
// whose name violates the convention, and its declaration comment metadata.
type Violation struct {
	Obj        types.Object
	DocComment astutil.DocCommentMetadata
}

// Violations returns the definitions of
// error-typed constants, variables, and named types whose name does not follow
// the [error naming] convention. Is is sorted by declaration position.
//
// [error naming]: https://go.dev/wiki/Errors#naming
func (c *Checker) Violations() []Violation {
	if c == nil {
		return nil
	}

	slices.SortFunc(c.violations, func(v1, v2 Violation) int { return cmp.Compare(v1.Obj.Pos(), v2.Obj.Pos()) })

	return c.violations
}

// CheckValueSpec walks the declarations in spec.
func (c *Checker) CheckValueSpec(spec *ast.ValueSpec, decl *ast.GenDecl) {
	for _, id := range spec.Names {
		def, ok := c.isNonConformantValue(id)
		if !ok {
			continue
		}

		doc, ok := astutil.CheckValueSpecComments(c.info, spec, decl, id, c.linter)
		if !ok {
			continue
		}

		c.violations = append(c.violations, Violation{Obj: def, DocComment: doc})
	}
}

// CheckTypeSpec checks the declaration in spec.
func (c *Checker) CheckTypeSpec(spec *ast.TypeSpec, decl *ast.GenDecl) {
	id := spec.Name

	tn, ok := c.isNonConformantType(id)
	if !ok {
		return
	}

	doc, ok := astutil.CheckTypeSpecComments(spec, decl, c.linter)
	if !ok {
		return
	}

	doc.TypeParams = spec.TypeParams
	c.violations = append(c.violations, Violation{Obj: tn, DocComment: doc})
}

// isNonConformantValue checks if a given identifier is a non-conformant variable/constant error object.
func (c *Checker) isNonConformantValue(id *ast.Ident) (types.Object, bool) {
	if id.Name == "_" {
		return nil, false
	}

	def, ok := c.info.Defs[id]
	if !ok || def == nil {
		return nil, false // should not happen
	}

	return def, typeutil.HasErrorMethodCached(&c.cache, def.Type()) && !conformantVar(def)
}

// isNonConformantType checks if a given identifier is a non-conformant variable/constant error object.
func (c *Checker) isNonConformantType(id *ast.Ident) (*types.TypeName, bool) {
	if id.Name == "_" {
		return nil, false
	}

	tn, ok := c.info.Defs[id].(*types.TypeName)
	if !ok {
		return nil, false // should not happen
	}

	// type definitions could have a value or pointer receiver
	return tn, typeutil.ErrorMethod.HasMethod(tn.Type(), true) && !conformantType(tn)
}
