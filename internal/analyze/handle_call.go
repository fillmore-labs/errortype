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

	"fillmore-labs.com/errortype/internal/typeutil"
)

// handleCall checks for incorrect pointer/value usage of error types passed to functions like errors.As.
func (p pass) handleCall(n *ast.CallExpr, opts AstOptions) {
	if len(n.Args) == 0 {
		return // Not interested in calls with no arguments.
	}

	fun, typeParams, methodExpr, ok := typeutil.FuncOf(p.TypesInfo, n.Fun)
	if !ok {
		return // Could not resolve function, might be a func variable.
	}

	funcName := typeutil.FuncNameOf(fun)

	info, ok := typeutil.KnownFuncs[funcName]
	if !ok {
		return
	}

	switch info.Kind() {
	case typeutil.KindIs:
		p.handleErrorIs(n, methodExpr, info.IsType(), opts.CheckIs)
		return

	case typeutil.KindEqu:
		p.handleErrorIs(n, methodExpr, info.IsType(), false)
		return

	case typeutil.KindAs:
		targetArgIndex, typeParam := info.AsTarget()

		// Handle generic functions like `errors.AsType[T]`.
		if typeParam >= 0 {
			if len(typeParams) > typeParam {
				typ := typeParams[typeParam]
				p.handleErrorsAsType(fun, typ)

				return
			}

			// Not enough type parameters were provided for the generic function.
			// If this is a generic-only function (no argument index fallback), we can't proceed.
			if targetArgIndex < 0 {
				return
			}
			// Otherwise, fall through to argument-based analysis.
		}

		if methodExpr {
			// For method expression calls ("(*assert.Assertions).ErrorsAs(a, ...)"),
			// the receiver `a` is the first argument. The argument indices in `errorsAs`
			// are for the function form, so we increment the index to correctly locate
			// the target argument in the method call expression.
			targetArgIndex++
		}

		p.handleErrorsAs(n, fun, targetArgIndex, opts)

		return
	}
}
