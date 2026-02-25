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

package wrappers

import (
	"go/ast"
	"go/types"

	"fillmore-labs.com/errortype/internal/detectwrappers/result"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// matchWrapperArgs determines if the arguments of a function call match the expected arguments for a wrapper.
func matchWrapperArgs(
	info *types.Info, fun typeutil.ResolvedFunc, args []ast.Expr, ef result.ErrorFunc,
	wantType result.WrapperType, srcVar, tgtVar *types.Var,
) bool {
	if ef.Type != wantType {
		return false
	}

	if src, ok := ef.SourceArg(args, fun.MethodExpr); !ok || !matchArg(info, src, srcVar) {
		return false
	}

	tgt, ok := ef.TargetArg(args, fun.MethodExpr)

	return ok && matchArg(info, tgt, tgtVar)
}

func matchWrapperType(
	info *types.Info, fun typeutil.ResolvedFunc, args []ast.Expr, ef result.ErrorFunc,
	srcVar *types.Var, tParam *types.TypeParam,
) bool {
	if ef.Type != result.WrapperAs && ef.Type != result.WrapperAsType {
		return false
	}

	if src, ok := ef.SourceArg(args, fun.MethodExpr); !ok || !matchArg(info, src, srcVar) {
		return false
	}

	targetType, ok := ef.TargetType(info, fun, args)

	return ok && types.Identical(tParam, targetType)
}

func matchArgs(info *types.Info, arg ast.Expr, srcVar, tgtVar *types.Var) bool {
	return matchArg(info, arg, srcVar) || tgtVar != nil && matchArg(info, arg, tgtVar)
}

func matchArg(info *types.Info, arg ast.Expr, v *types.Var) bool {
	id, ok := ast.Unparen(arg).(*ast.Ident)
	if !ok || id.Name != v.Name() {
		return false
	}

	argV, ok := info.Uses[id].(*types.Var)
	if !ok {
		return false
	}

	return argV == v
}
