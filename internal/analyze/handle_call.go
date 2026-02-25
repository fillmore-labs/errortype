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

	"fillmore-labs.com/errortype/internal/detectwrappers/result"
	"fillmore-labs.com/errortype/internal/errorconv"
	"fillmore-labs.com/errortype/internal/errorlike"
	"fillmore-labs.com/errortype/internal/settings"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// handleCall checks for incorrect pointer/value usage of error types passed to error handling functions.
func (p Pass) handleCall(c inspector.Cursor, call *ast.CallExpr) {
	if len(call.Args) == 0 {
		return // Not interested in calls with no arguments.
	}

	for argType, argExpr := range errorconv.CallArgs(p.TypesInfo, call) {
		p.checkCallArg(argType, argExpr)
	}

	fun, wrapper, ok := p.Lookup(p.TypesInfo, call)
	if !ok {
		return // Not a known error wrapper.
	}

	var unusedCandidate, checkRecv bool

	switch wrapper.Type {
	case result.WrapperIs:
		src, ok := wrapper.SourceArg(call.Args, fun.MethodExpr)
		if !ok {
			break // multi-valued argument
		}

		tgt, ok := wrapper.TargetArg(call.Args, fun.MethodExpr)
		if !ok {
			break // multi-valued argument
		}

		p.comparison(call, src, tgt, false)

		unusedCandidate, checkRecv = true, true

	case result.WrapperAs:
		if tgt, ok := wrapper.TargetArg(call.Args, fun.MethodExpr); ok {
			p.handleErrorsAs(call, fun.Expr, tgt)
		}

		// Flag ignored return values even though errors.As populates target:
		// checking `target != nil` instead of the returned bool is poor style
		// and misses typed-nil errors (where errors.As returns true with target == nil).
		unusedCandidate, checkRecv = true, true

	case result.WrapperAsType:
		// Type argument checking is handled once in handleInstances.
		unusedCandidate = true

	case result.WrapperUnwrap:
		unusedCandidate = true

	case result.WrapperErrorf:
		unusedCandidate, checkRecv = true, true

		if call.Ellipsis.IsValid() {
			break // The operands are passed as a slice and can not be matched to verbs.
		}

		formatExpr, ok := wrapper.SourceArg(call.Args, fun.MethodExpr)

		operands := wrapper.TargetArgs(call.Args, fun.MethodExpr)
		if !ok || len(operands) == 0 {
			break // multi-valued argument or no operands
		}

		p.handleErrorf(formatExpr, operands)

	case result.FuncIsType:
		if typeExpr, ok := wrapper.SourceArg(call.Args, fun.MethodExpr); ok {
			if targetExpr, ok := wrapper.TargetArg(call.Args, fun.MethodExpr); ok {
				p.handleIsType(typeExpr, targetExpr)
			}
		}

		// Currently, these are all test assertions
		unusedCandidate = false

	case result.FuncEqual:
		if left, ok := wrapper.SourceArg(call.Args, fun.MethodExpr); ok {
			if right, ok := wrapper.TargetArg(call.Args, fun.MethodExpr); ok {
				p.comparison(call, left, right, true)
			}
		}

		// Currently, these are all test assertions
		unusedCandidate = false

	case result.FuncAssert:
		typ, ok := wrapper.TargetTypeArg(p.TypesInfo, fun.Ident)
		if !ok { // should not happen
			return
		}

		if !errorlike.Implements(typ) {
			return // we are only interested in assertions to error types.
		}

		// Check if the error type is used correctly (pointer vs. value).
		p.checkGenericCall(typ, call, fun.Expr)
		unusedCandidate = true

	default:
		p.ReportErrorf(fun.Expr, "Unexpected function type: %v", wrapper.Type)
	}

	// Report ignored results for wrappers whose return value is expected to be used.
	if unusedCandidate && p.Has(settings.OptionCheckUnused) {
		// checkRecv is a crude heuristic: When the source is not the first argument, it could be some kind of helper.
		if checkRecv && wrapper.Source != 0 {
			return
		}

		p.checkUnused(c, call, fun, checkRecv, wrapper.Type == result.WrapperIs)
	}
}

func (p Pass) checkUnused(c inspector.Cursor, call *ast.CallExpr, fun typeutil.ResolvedFunc, checkRecv, plus bool) {
	sig := typeutil.SignatureOf(fun.Func)
	if sig == nil || sig.Results().Len() == 0 {
		return
	}

	// checkRecv is a crude heuristic: When there is a receiver, it could be some kind of helper.
	if checkRecv && sig.Recv() != nil {
		return
	}

	e := c

unwrap:
	switch e = e.Parent(); e.Node().(type) {
	case *ast.ParenExpr:
		goto unwrap

	case *ast.ExprStmt:
		tag := "unu"
		if plus {
			tag = "unu+"
		}

		p.ReportRangef(call, "Result of %s is ignored (et:%s)", exprToString(p.Fset, call.Fun), tag)
	}
}
