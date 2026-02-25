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
	"slices"

	"fillmore-labs.com/errortype/internal/naming/rules"
)

// mapUsesByObject maps each violation to its use-site positions and the set of
// violation types that are embedded by name in a struct.
func mapUsesByObject(info *types.Info, violations []rules.Violation) (map[types.Object][]token.Pos, embeddedTypes) {
	// Precompute object uses, indexed by renamed object.
	usesByObj := make(map[types.Object][]token.Pos, len(violations))
	for _, v := range violations {
		usesByObj[v.Obj] = nil
	}

	var embeddings embeddedTypes
	// Collect use positions and discover embedding sites in one pass.
	for id, use := range info.Uses {
		uses, ok := usesByObj[use]
		if !ok {
			continue // not a use we are interested in
		}

		// Check for embedded types. Assert we have a [*types.TypeName] first,
		// since its faster than a map lookup.
		if tn, ok := use.(*types.TypeName); ok && isEmbedded(info, id) {
			// Alias embedded types rather than rewriting the embedding site: it
			// is rare, and a rename there would force us to follow a chain when
			// the embedding struct is itself embedded.
			embeddings.add(tn)

			continue // leave the embedding site untouched
		}

		usesByObj[use] = append(uses, id.Pos())
	}

	// [types.Info.Uses] is a map and in random order. We later look up
	// files with [token.FileSet.File], which caches the last looked up file.
	for _, uses := range usesByObj {
		slices.Sort(uses)
	}

	return usesByObj, embeddings
}

func isEmbedded(info *types.Info, id *ast.Ident) bool {
	// An embedded field's identifier is the only one that appears in both
	// Uses (the embedded *TypeName) and Defs (the synthesized field *Var).
	def, ok := info.Defs[id]
	if !ok {
		return false
	}

	// invariant check
	if v, ok := def.(*types.Var); !ok || !v.Embedded() {
		panic(fmt.Errorf("internal error: not an embedded field: %#v", def))
	}

	return true
}

type embeddedTypes map[*types.TypeName]struct{}

// add marks a type as embedded.
func (e *embeddedTypes) add(tn *types.TypeName) {
	if *e == nil {
		*e = make(embeddedTypes)
	}

	(*e)[tn] = struct{}{}
}

// isEmbedded reports whether obj is a type embedded by name in a struct,
// as collected by [mapUsesByObject].
func (e embeddedTypes) isEmbedded(obj types.Object) bool {
	tn, ok := obj.(*types.TypeName)
	if !ok {
		return false
	}

	_, ok = e[tn]

	return ok
}
