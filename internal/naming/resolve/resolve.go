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

package resolve

import (
	"go/ast"
	"go/token"
	"go/types"

	"fillmore-labs.com/errortype/internal/naming/rules"
)

// Resolver bundles the context needed in [Resolver.Name] to decide
// whether a proposed name conflicts with existing declarations or claimed names.
type Resolver struct {
	// claims is the set of new names chosen by fixable renames.
	claims     map[string]struct{}
	pkgScope   *types.Scope
	fileScopes []*types.Scope
}

// New creates and initializes a Resolver instance for managing proposed renames and conflict resolution.
func New(info *types.Info, pkgScope *types.Scope, files []*ast.File, numRenames int) *Resolver {
	fileScopes := collectFileScopes(info, files)

	return &Resolver{
		claims:     make(map[string]struct{}, numRenames),
		pkgScope:   pkgScope,
		fileScopes: fileScopes,
	}
}

// Name chooses a conflict-free name. It starts from the [rules.Suggest] proposal
// and, on conflict, tries numbered variants; if all are exhausted the rename is returned non-fixable.
//
// A candidate that already arrived non-fixable keeps the bare suggestion for
// display: it claims no name, so there is no point resolving conflicts for it.
func (r *Resolver) Name(obj types.Object, fixable bool, uses []token.Pos) (string, bool) {
	suggestion := rules.Suggest(obj)
	newName := suggestion.Name()

	if !fixable || newName == "" {
		return newName, false
	}

	if !r.hasConflict(newName, uses) {
		r.claim(newName)
		return newName, true
	}

	// After the base name, variants "errX2" through "errX9" (or "x2Error" through
	// "x9Error" for types) are tried. Renumbering is rare, so re-scanning the enclosing
	// scope per attempt is cheaper than precomputing for the common conflict-free case.
	const maxVariants = 8
	for i := range maxVariants {
		if name := suggestion.Numbered(i + 2); !r.hasConflict(name, uses) {
			r.claim(name)
			return name, true
		}
	}

	return newName, false
}

// collectFileScopes returns the file-block scopes of the package being checked.
func collectFileScopes(info *types.Info, files []*ast.File) []*types.Scope {
	scopes := make([]*types.Scope, 0, len(files))

	for _, file := range files {
		if scope := info.Scopes[file]; scope != nil {
			scopes = append(scopes, scope)
		}
	}

	return scopes
}
