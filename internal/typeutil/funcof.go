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

// ResolvedFunc is the result of resolving a function expression.
type ResolvedFunc struct {
	Ident      *ast.Ident
	Func       *types.Func
	Expr       ast.Expr
	MethodExpr bool
}

// FuncOf iteratively unwraps an expression to find the underlying function declaration.
func FuncOf(info *types.Info, call *ast.CallExpr) (res ResolvedFunc, ok bool) {
	res.Expr = call.Fun
	typeParams := false

unwarp:
	switch e := res.Expr.(type) {
	case *ast.Ident:
		res.Ident = e
		res.Func, ok = info.Uses[e].(*types.Func)

		return res, ok

	case *ast.SelectorExpr:
		if sel, isSel := info.Selections[e]; isSel {
			switch sel.Kind() {
			case types.MethodVal:
				res.MethodExpr = false

			case types.MethodExpr:
				res.MethodExpr = true

			default: // types.FieldVal, struct field selector
				return res, false
			}

			res.Ident = e.Sel
			res.Func = sel.Obj().(*types.Func)

			return res, true
		}

		res.Ident = e.Sel
		res.Func, ok = info.Uses[e.Sel].(*types.Func) // e.Sel is an identifier qualified by e.X

		return res, ok

	case *ast.IndexExpr: // Generic function instantiation with a type parameter ("myFunc[T]").
		if typeParams { // should not happen, duplicate type parameters
			return res, false
		}

		if !checkTypeParameters(info, []ast.Expr{e.Index}) {
			return res, false // No type, but an array/slice index.
		}

		typeParams = true
		res.Expr = e.X // Unwrap to the function identifier.

		goto unwarp

	case *ast.IndexListExpr: // Generic function instantiation with multiple type parameters ("myFunc[T, U]").
		if typeParams { // should not happen, duplicate type parameters
			return res, false
		}

		if !checkTypeParameters(info, e.Indices) { // should not happen
			return res, false
		}

		typeParams = true
		res.Expr = e.X // Unwrap to the function identifier.

		goto unwarp

	case *ast.ParenExpr: // Parenthesized expression ("(myFunc)")
		res.Expr = e.X // Unwrap to the inner expression.
		goto unwarp

	default: // Function variable, pointer, or other non-declarative function reference.
		return res, false
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
