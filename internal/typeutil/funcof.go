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

package typeutil

import (
	"go/ast"
	"go/types"
)

// FuncOf iteratively unwraps an expression to find the underlying function declaration.
func FuncOf(info *types.Info, n *ast.CallExpr) (fun *types.Func, typeParams []ast.Expr, methodExpr, ok bool) {
	for ex := n.Fun; ; {
		switch e := ex.(type) {
		case *ast.Ident:
			fun, ok = info.Uses[e].(*types.Func)

			return fun, typeParams, false, ok

		case *ast.SelectorExpr:
			if sel, isSel := info.Selections[e]; isSel {
				switch sel.Kind() {
				case types.MethodVal:
					methodExpr = false

				case types.MethodExpr:
					methodExpr = true

				default: // types.FieldVal, struct field selector
					return nil, nil, false, false
				}

				fun = sel.Obj().(*types.Func)

				return fun, typeParams, methodExpr, true
			}

			if fun, ok = info.Uses[e.Sel].(*types.Func); ok { // e.Sel is an identifier qualified by e.X
				return fun, typeParams, false, true
			}

			return nil, nil, false, false // Not a function

		case *ast.IndexExpr: // Generic function instantiation with a type parameter ("myFunc[T]").
			if typeParams != nil { // should not happen, duplicate type parameters
				return nil, nil, false, false
			}

			typeParams = []ast.Expr{e.Index}
			if !checkTypeParameters(info, typeParams) {
				return nil, nil, false, false // No type, but an array/slice index.
			}

			ex = e.X // Unwrap to the function identifier.

		case *ast.IndexListExpr: // Generic function instantiation with multiple type parameters ("myFunc[T, U]").
			if typeParams != nil { // should not happen, duplicate type parameters
				return nil, nil, false, false
			}

			typeParams = e.Indices
			if !checkTypeParameters(info, typeParams) { // should not happen
				return nil, nil, false, false
			}

			ex = e.X // Unwrap to the function identifier.

		case *ast.ParenExpr: // Parenthesized expression ("(myFunc)")
			ex = e.X // Unwrap to the inner expression.

		default: // Function variable, pointer, or other non-declarative function reference.
			return nil, nil, false, false
		}
	}
}

// checkTypeParameters validates type parameters from an IndexExpr
// or IndexListExpr. It uses the provided types.Info to verify that each
// expression represents a type.
//
// If any expression is not a type, it returns false.
func checkTypeParameters(info *types.Info, indices []ast.Expr) bool {
	for _, index := range indices {
		if !info.Types[index].IsType() { // Must be a type parameter, not an array/slice index.
			return false
		}
	}

	return true
}
