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

	"golang.org/x/tools/go/ast/inspector"

	"fillmore-labs.com/errortype/internal/knownfuncs"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// handleCall checks for incorrect pointer/value usage of error types passed to functions like errors.As.
func (p Pass) handleCall(c inspector.Cursor, n *ast.CallExpr) {
	if len(n.Args) == 0 {
		return // Not interested in calls with no arguments.
	}

	fun, typeParams, methodExpr, ok := typeutil.FuncOf(p.TypesInfo, n)
	if !ok {
		return // Could not resolve function, might be a func variable.
	}

	info, ok := knownfuncs.FuncInfoOf(fun)
	if !ok {
		return
	}

	// TODO: Handle deprecation

	switch info.Kind() {
	case knownfuncs.KindIs:
		p.handleErrorIs(n, methodExpr, info.IsType(), false)

	case knownfuncs.KindEqu:
		p.handleErrorIs(n, methodExpr, info.IsType(), true)

	case knownfuncs.KindAs:
		targetArgIndex, typeParam := info.AsTarget()

		// Handle generic functions like `errors.AsType[T]`.
		// Check whether enough type parameters were provided for the generic function.
		if typeParam >= 0 && len(typeParams) > typeParam {
			typ := typeParams[typeParam]
			p.handleErrorsAsType(fun, typ)

			break
		}

		// Argument-based analysis.

		// If this is a generic-only function (no argument index fallback), we can't proceed.
		if targetArgIndex < 0 {
			break
		}

		if methodExpr {
			// For method expression calls ("(*assert.Assertions).ErrorsAs(a, ...)"),
			// the receiver `a` is the first argument. The argument indices in `errorsAs`
			// are for the function form, so we increment the index to correctly locate
			// the target argument in the method call expression.
			targetArgIndex++
		}

		p.handleErrorsAs(n, fun, targetArgIndex)

	case knownfuncs.KindType:
		p.handleIsType(n, methodExpr, info.IsType())
	}

	if tag, shouldCheck := checkResultWithTag(info.EvalType(), p.CheckUnused()); shouldCheck {
		for e := c.Parent(); ; e = e.Parent() {
			switch e.Node().(type) {
			case *ast.ParenExpr:
				continue

			case *ast.ExprStmt:
				p.ReportRangef(n, "Result of %s is ignored (et:%s)", fun.FullName(), tag)
			}

			break
		}
	}
}

// checkResultWithTag determines a tag and status based on the evaluation type and unused check flag.
func checkResultWithTag(evalType knownfuncs.EvalType, checkUnused bool) (string, bool) {
	switch evalType {
	case knownfuncs.MustEval:
		return "unu+", true

	case knownfuncs.ShouldEval:
		if !checkUnused {
			return "", false
		}

		return "unu", true

	default:
		return "", false
	}
}
