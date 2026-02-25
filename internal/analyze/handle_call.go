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
	"go/types"

	"golang.org/x/tools/go/ast/inspector"

	"fillmore-labs.com/errortype/detect/result"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// handleCall checks for incorrect pointer/value usage of error types passed to functions like errors.As.
func (p Pass) handleCall(c inspector.Cursor, call *ast.CallExpr) {
	if len(call.Args) == 0 {
		return // Not interested in calls with no arguments.
	}

	fun, ok := typeutil.FuncOf(p.TypesInfo, call)
	if !ok {
		return // Could not resolve function, might be a field.
	}

	wrapper, ok := p.ErrorFuncs[fun.Func]
	if !ok {
		return // Not a function we are interested in.
	}

	var unusedCandidate bool

	switch wrapper.Type {
	case result.WrapperIs:
		if len(call.Args) < 2 {
			break // multi-valued argument
		}

		srcIdx, tgtIdx := int(wrapper.Source), int(wrapper.Target)
		if fun.MethodExpr {
			srcIdx++
			tgtIdx++
		}

		p.comparison(call, call.Args[srcIdx], call.Args[tgtIdx], false)

		unusedCandidate = expectResultUsed(wrapper, fun)

	case result.WrapperAs:
		tgtIdx := int(wrapper.Target)
		if fun.MethodExpr {
			tgtIdx++
		}

		p.handleErrorsAs(call, fun.Expr, tgtIdx)

		// errors.As is used without checking the return value.
		// This does not account for nil errors, which should be caught even though they are bad style.
		//
		//	var err *net.OpError
		//	var target net.Error
		//	if errors.As(err, &target) { /* ... */ }
		unusedCandidate = expectResultUsed(wrapper, fun)

	case result.WrapperAsType:
		instance, ok := p.TypesInfo.Instances[fun.Ident]
		if !ok { // should not happen
			return
		}

		tgtIdx := int(wrapper.Target)
		typ := instance.TypeArgs.At(tgtIdx)

		// Check if the error type is used correctly (pointer vs. value).
		p.checkGenericCall(typ, call, fun.Expr)
		unusedCandidate = true

	case result.WrapperErrorf:
		unusedCandidate = true

		if len(call.Args) < 2 {
			break // multi-valued argument
		}

		srcIdx, tgtIdx := int(wrapper.Source), int(wrapper.Target)
		if fun.MethodExpr {
			srcIdx++
			tgtIdx++
		}

		p.handleErrorf(call, srcIdx, tgtIdx)

	case result.FuncIsType:
		p.handleIsType(call, fun.MethodExpr, wrapper.Source)

		// Currently, these are all test assertions
		unusedCandidate = false

	case result.FuncEqual:
		argIndex := int(wrapper.Source)
		if fun.MethodExpr {
			argIndex++
		}

		p.handleEqual(call, argIndex)

		// Currently, these are all test assertions
		unusedCandidate = false

	case result.FuncAssert:
		instance, ok := p.TypesInfo.Instances[fun.Ident]
		if !ok { // should not happen
			return
		}

		typ := instance.TypeArgs.At(int(wrapper.Target))
		if !typeutil.HasErrorMethod(typ) {
			return // we are only interested in assertions to error types.
		}

		// Check if the error type is used correctly (pointer vs. value).
		p.checkGenericCall(typ, call, fun.Expr)
		unusedCandidate = true

	default:
		p.ReportErrorf(fun.Expr, "Unexpected function type: %v", wrapper.Type)
	}

	// For Is/As/AsType wrappers, it's reasonable to expect the result to be used.

	var sig *types.Signature
	if unusedCandidate && p.Has(OptionCheckUnused) {
		sig = typeutil.SignatureOf(fun.Func)
	}

	if sig != nil && sig.Results().Len() > 0 {
		tag := "unu"
		if wrapper.Type == result.WrapperIs {
			tag = "unu+"
		}

		p.checkUnused(c, call, tag)
	}
}

// expectResultUsed is a crude heuristic: When there is a parameter
// before the error or a receiver, it could be a test helper.
func expectResultUsed(wrapper result.ErrorFunc, fun typeutil.ResolvedFunc) bool {
	if wrapper.Source != 0 {
		return false
	}

	sig := typeutil.SignatureOf(fun.Func)

	return sig != nil && sig.Recv() == nil
}

func (p Pass) checkUnused(c inspector.Cursor, call *ast.CallExpr, tag string) {
	e := c.Parent()

unwrap:
	switch e.Node().(type) {
	case *ast.ParenExpr:
		e = e.Parent()
		goto unwrap

	case *ast.ExprStmt:
		p.ReportRangef(call, "Result of %s is ignored (et:%s)", exprToString(p.Fset, call.Fun), tag)
	}
}
