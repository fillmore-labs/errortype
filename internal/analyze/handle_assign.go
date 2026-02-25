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

package analyze

import (
	"go/ast"
	"go/token"
	"go/types"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// handleAssign checks assignments of concrete error types to variables
// with an error-like interface type.
func (p Pass) handleAssign(stmt *ast.AssignStmt) {
	var shortDecl bool

	switch stmt.Tok {
	case token.ASSIGN:

	case token.DEFINE:
		shortDecl = true

	default:
		return
	}

	for idx, lhs := range stmt.Lhs {
		var lTyp types.Type

		switch shortDecl {
		case true:
			id, ok := lhs.(*ast.Ident)
			if !ok || id.Name == "_" {
				continue
			}

			v, ok := p.TypesInfo.Uses[id]

			if !ok { // We have a newly declared variable, handle it as a variable declaration.
				if p.PrefixFilter() && !hasErrPrefix(id.Name) {
					continue
				}

				rTyp, expr := typeutil.ResultOf(p.TypesInfo, stmt.Rhs, idx)
				if rTyp == nil || !typeutil.HasErrorMethod(rTyp) {
					continue // Not an error type
				}

				p.checkVarDecl(rTyp, expr, id.Name)

				continue // Skip the assignment interface target check
			}

			lTyp = v.Type() // continue with assignment

		default:
			tv, ok := p.TypesInfo.Types[lhs]
			if !ok {
				continue // blank identifier
			}

			lTyp = tv.Type
		}

		if !typeutil.IsInterfaceWithError(lTyp) {
			continue
		}

		rTyp, expr := typeutil.ResultOf(p.TypesInfo, stmt.Rhs, idx)
		if rTyp == nil { // should not happen
			continue
		}

		p.checkAssign(rTyp, expr)
	}
}
