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

package result

import (
	"go/ast"
	"go/types"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// Lookup resolves the function called by call and returns its [ErrorFunc] metadata if known.
func (e ErrorFuncs) Lookup(info *types.Info, call *ast.CallExpr) (typeutil.ResolvedFunc, ErrorFunc, bool) {
	fun, ok := typeutil.FuncOf(info, call)
	if !ok {
		return fun, ErrorFunc{}, false
	}

	ef, ok := e[fun.Func]

	return fun, ef, ok
}

// IsAsTypeTarget reports whether the type argument at index of obj is the
// target type parameter of a detected [WrapperAsType] wrapper.
func (e ErrorFuncs) IsAsTypeTarget(obj types.Object, index int) bool {
	if fn, ok := obj.(*types.Func); ok {
		obj = fn.Origin()
	}

	ef, ok := e[obj]

	return ok && ef.Type == WrapperAsType && int(ef.Target) == index
}

// SourceArg returns the source argument expression from args, accounting for
// method-expression receiver shifts. It returns false when the wrapper has no
// source argument (such as [FuncAssert]) or args has too few elements (e.g., a
// multi-valued call argument).
func (w ErrorFunc) SourceArg(args []ast.Expr, methodExpr bool) (ast.Expr, bool) {
	return argAt(args, w.Source, methodExpr)
}

// TargetArg returns the target value argument expression from args, accounting
// for method-expression receiver shifts. It returns false for wrappers whose
// target is a type parameter ([WrapperAsType], [FuncAssert]) or when args has
// too few elements.
func (w ErrorFunc) TargetArg(args []ast.Expr, methodExpr bool) (ast.Expr, bool) {
	switch w.Type {
	case WrapperIs, WrapperAs, WrapperErrorf, FuncIsType, FuncEqual:
		return argAt(args, w.Target, methodExpr)

	default:
		return nil, false
	}
}

// TargetArgs returns the slice of target arguments starting at [ErrorFunc.Target]
// for variadic wrappers ([WrapperErrorf]), accounting for method-expression
// receiver shifts. It returns nil if the wrapper is not variadic or has no
// target operands.
func (w ErrorFunc) TargetArgs(args []ast.Expr, methodExpr bool) []ast.Expr {
	if w.Type != WrapperErrorf || w.Target < 0 {
		return nil
	}

	i := int(w.Target)
	if methodExpr {
		i++
	}

	if i >= len(args) {
		return nil
	}

	return args[i:]
}

// TargetTypeArg returns the target type argument for generic wrappers
// ([WrapperAsType], [FuncAssert]) from info.Instances.
func (w ErrorFunc) TargetTypeArg(info *types.Info, ident *ast.Ident) (types.Type, bool) {
	switch w.Type {
	case WrapperAsType, FuncAssert:
	default:
		return nil, false
	}

	if w.Target < 0 {
		return nil, false
	}

	instance, ok := info.Instances[ident]
	if !ok || int(w.Target) >= instance.TypeArgs.Len() {
		return nil, false
	}

	return instance.TypeArgs.At(int(w.Target)), true
}

// TargetType extracts the target error type checked by a call to this wrapper:
//   - [WrapperIs]: the type of the target argument.
//   - [WrapperAs]: the element type of the target pointer argument.
//   - [WrapperAsType], [FuncAssert]: the target type argument.
//
// It returns false for wrappers without a target error type ([WrapperErrorf],
// [FuncIsType], [FuncEqual]) or if the target cannot be resolved.
func (w ErrorFunc) TargetType(info *types.Info, fun typeutil.ResolvedFunc, args []ast.Expr) (types.Type, bool) {
	switch w.Type {
	case WrapperIs:
		arg, ok := w.TargetArg(args, fun.MethodExpr)
		if !ok {
			return nil, false
		}

		tv, ok := info.Types[arg]
		if !ok || tv.Type == nil {
			return nil, false
		}

		return tv.Type, true

	case WrapperAs:
		arg, ok := w.TargetArg(args, fun.MethodExpr)
		if !ok {
			return nil, false
		}

		tv, ok := info.Types[arg]
		if !ok || tv.Type == nil {
			return nil, false
		}

		ptr, ok := tv.Type.Underlying().(*types.Pointer)
		if !ok {
			return nil, false
		}

		return ptr.Elem(), true

	case WrapperAsType, FuncAssert:
		return w.TargetTypeArg(info, fun.Ident)

	default:
		return nil, false
	}
}

// argAt returns args[idx] (shifted by 1 for method expressions), or false if
// idx is negative or out of bounds.
//
// Unlike [typeutil.ResultOf], it intentionally does not unwrap multi-valued
// calls (`f(g())`, where args has length 1 and a *[types.Tuple] type): Go
// requires g() to supply all arguments of f, which offers little benefit here
// (it would have to return both sides of a comparison, a non-constant format
// string, a target pointer for errors.As without a caller variable, or even
// *testing.T alongside test assertion arguments).
func argAt(args []ast.Expr, idx int8, methodExpr bool) (ast.Expr, bool) {
	if idx < 0 {
		return nil, false
	}

	i := int(idx)
	if methodExpr {
		i++
	}

	if i >= len(args) {
		return nil, false
	}

	return args[i], true
}
