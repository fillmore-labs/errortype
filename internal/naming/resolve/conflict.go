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
	"go/token"
	"go/types"
)

// hasConflict reports whether declaring newName at package scope would clash.
// The checks, in order, are:
//
//   - the name is already declared or claimed at package scope;
//   - the name is declared or claimed in an inner scope between a use site and
//     package scope, which would shadow that use of the renamed object;
//   - the name is a file-block name such as an import.
func (r *Resolver) hasConflict(newName string, uses []token.Pos) bool {
	// Check if newName is already declared or claimed in the declaration scope.
	if r.pkgScope.Lookup(newName) != nil {
		return true
	}

	if r.claimed(newName) {
		return true
	}

	checked := make(map[*types.Scope]struct{})
	checked[r.pkgScope] = struct{}{}

	// Check if newName is declared or claimed in any inner scope between a
	// use site and package scope.
	for _, pos := range uses {
		for scope := r.pkgScope.Innermost(pos); scope != nil; scope = scope.Parent() {
			if _, ok := checked[scope]; ok {
				break
			}

			if scope.Lookup(newName) != nil {
				return true
			}

			checked[scope] = struct{}{}
		}
	}

	// Package-level declarations also collide with file-block names like imports.
	// "no identifier may be declared in both the file and package block"
	// https://go.dev/ref/spec#Declarations_and_scope
	for _, fileScope := range r.fileScopes {
		if fileScope.Lookup(newName) != nil {
			return true
		}
	}

	return false
}

func (r *Resolver) claim(name string) {
	r.claims[name] = struct{}{}
}

func (r *Resolver) claimed(name string) bool {
	_, ok := r.claims[name]
	return ok
}
