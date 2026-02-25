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
	Func       types.Object
	Expr       ast.Expr
	MethodExpr bool
}

// FuncOf iteratively unwraps an expression to find the underlying function declaration.
func FuncOf(info *types.Info, call *ast.CallExpr) (res ResolvedFunc, ok bool) {
	res.Expr = call.Fun
	typeParams := false

unwarp:
	switch expr := res.Expr.(type) {
	case *ast.Ident:
		return returnIdent(info, expr, res)

	case *ast.SelectorExpr:
		if sel, ok := info.Selections[expr]; ok {
			// selector expression
			res.Ident = expr.Sel

			switch sel.Kind() {
			case types.MethodVal:

			case types.MethodExpr:
				res.MethodExpr = true

			default: // types.FieldVal, struct field selector
				return res, false
			}

			res.Func = sel.Obj().(*types.Func).Origin()

			return res, true
		}

		// qualified identifier, e.X is a package name
		return returnIdent(info, expr.Sel, res)

	case *ast.IndexExpr: // Generic function instantiation with a type parameter ("myFunc[T]").
		if typeParams { // should not happen, duplicate type parameters
			return res, false
		}

		if !checkTypeParameters(info, []ast.Expr{expr.Index}) {
			return res, false // No type, but an array/slice index.
		}

		typeParams = true
		res.Expr = expr.X // Unwrap to the function identifier.

		goto unwarp

	case *ast.IndexListExpr: // Generic function instantiation with multiple type parameters ("myFunc[T, U]").
		if typeParams { // should not happen, duplicate type parameters
			return res, false
		}

		if !checkTypeParameters(info, expr.Indices) { // should not happen
			return res, false
		}

		typeParams = true
		res.Expr = expr.X // Unwrap to the function identifier.

		goto unwarp

	case *ast.ParenExpr: // Parenthesized expression ("(myFunc)")
		res.Expr = expr.X // Unwrap to the inner expression.

		goto unwarp

	default: // Pointer or other non-declarative function reference.
		return res, false
	}
}

func returnIdent(info *types.Info, id *ast.Ident, res ResolvedFunc) (ResolvedFunc, bool) {
	res.Ident = id

	fun, ok := info.Uses[id]
	if !ok {
		return res, false
	}

	switch fun := fun.(type) {
	case *types.Func:
		res.Func = fun

		return res, true

	case *types.Var:
		if !PackageLevel(fun) {
			return res, false
		}

		res.Func = fun

		return res, true

	default:
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
