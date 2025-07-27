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
func FuncOf(info *types.Info, ex ast.Expr) (fun *types.Func, typeParams []ast.Expr, methodExpr, ok bool) {
	var tp []ast.Expr

	for {
		switch e := ex.(type) {
		case *ast.Ident:
			fun, ok = info.Uses[e].(*types.Func)

			return fun, tp, false, ok

		case *ast.SelectorExpr:
			sel, ok := info.Selections[e]
			if !ok { // e.Sel is an identifier qualified by e.X
				fun, ok = info.Uses[e.Sel].(*types.Func) // types.Checker calls recordUse for e.Sel from recordSelection.

				return fun, tp, false, ok
			}

			switch sel.Kind() { //nolint:exhaustive
			case types.MethodVal: // e.Sel is a method selector
				fun, ok = sel.Obj().(*types.Func)

				return fun, tp, false, ok

			case types.MethodExpr: // e.Sel is a method expression
				fun, ok = sel.Obj().(*types.Func)

				return fun, tp, true, ok
			}

			return nil, nil, false, false // e.Sel is a struct field selector

		case *ast.IndexExpr: // Generic function instantiation with a type parameter ("myFunc[T]").
			if len(tp) > 0 { // Duplicate type parameters, shouldn't happen
				return nil, nil, false, false
			}

			typeParam := info.Types[e.Index]
			if !typeParam.IsType() {
				return nil, nil, false, false // Must be a type parameter, not an array/slice index.
			}

			tp = []ast.Expr{e.Index}

			ex = e.X // Unwrap to the function identifier.

		case *ast.IndexListExpr: // Generic function instantiation with multiple type parameters ("myFunc[T, U]").
			if len(tp) > 0 { // Duplicate type parameters, shouldn't happen
				return nil, nil, false, false
			}

			tp = make([]ast.Expr, 0, len(e.Indices))
			for _, index := range e.Indices {
				typeParam := info.Types[index]
				if !typeParam.IsType() {
					return nil, nil, false, false // Must be a type parameter, not an array/slice index.
				}

				tp = append(tp, index)
			}

			ex = e.X // Unwrap to the function identifier.

		case *ast.ParenExpr: // Parenthesized expression ("(myFunc)")
			ex = e.X // Unwrap to the inner expression.

		default: // The expression does not resolve to a function identifier (could be a function pointer).
			return nil, nil, false, false
		}
	}
}
