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
	"go/token"
	"go/types"

	"fillmore-labs.com/errortype/detect/result"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// resolveCandidate traverses the AST of a wrapper candidate to find internal calls to known wrappers or other candidates.
func resolveCandidate(info *types.Info, cand *wrapperCandidate, wrapperCandidates map[*types.Func]*wrapperCandidate, known, wrappers result.ErrorFuncs) {
	// We do not use [golang.org/x/tools/go/ast/inspector.Inspector] here since we assume most packages don't contain wrappers.
	for n := range ast.Preorder(cand.body) {
		switch n := n.(type) {
		case *ast.CallExpr:
			if matchCall(info, cand, n, wrapperCandidates, known, wrappers) {
				propagate(wrappers, cand)
				return // first matching call only
			}

		case *ast.AssignStmt:
			for _, lhs := range n.Lhs {
				if matchArgs(info, lhs, cand.srcVar, cand.tgtVar) {
					return // Argument is modified
				}
			}

		case *ast.UnaryExpr:
			if n.Op == token.AND {
				if matchArgs(info, n.X, cand.srcVar, cand.tgtVar) {
					return // Argument has address taken
				}
			}
		}
	}
}

func matchCall(info *types.Info, cand *wrapperCandidate, call *ast.CallExpr, wrapperCandidates map[*types.Func]*wrapperCandidate, known, wrappers result.ErrorFuncs) bool {
	fun, ok := typeutil.FuncOf(info, call)
	if !ok {
		return false // Could not resolve the function, might be a func variable.
	}

	// 1. Is it a call to an already known wrapper?
	ef, ok := known[fun.Func]
	if !ok {
		ef, ok = wrappers[fun.Func]
	}

	if ok {
		return checkKnownWrapper(info, fun, call.Args, ef, cand)
	}

	// 2. Is it a call to another wrapperCandidate wrapper?
	if callee, ok := wrapperCandidates[fun.Func]; ok {
		registerCaller(info, fun, call.Args, callee, cand)
	}

	return false
}

func checkKnownWrapper(info *types.Info, fun typeutil.ResolvedFunc, args []ast.Expr, ef result.ErrorFunc, cand *wrapperCandidate) bool {
	switch typ := cand.errorFunc.Type; typ {
	case result.WrapperIs, result.WrapperAs:
		return matchIsAs(info, fun, args, ef, typ, cand.srcVar, cand.tgtVar)

	case result.WrapperAsType:
		return matchAsType(info, fun, args, ef, cand.srcVar, cand.tParam)
	}

	return false
}

func registerCaller(info *types.Info, fun typeutil.ResolvedFunc, args []ast.Expr, callee, cand *wrapperCandidate) {
	switch typ := cand.errorFunc.Type; typ {
	case result.WrapperIs, result.WrapperAs:
		if matchIsAs(info, fun, args, callee.errorFunc, typ, cand.srcVar, cand.tgtVar) {
			callee.callers = append(callee.callers, cand)
		}

	case result.WrapperAsType:
		if matchAsType(info, fun, args, callee.errorFunc, cand.srcVar, cand.tParam) {
			callee.callers = append(callee.callers, cand)
		}
	}
}

func propagate(wrappers result.ErrorFuncs, cand *wrapperCandidate) {
	if _, ok := wrappers[cand.fun]; ok {
		return // already propagated
	}

	wrappers[cand.fun] = cand.errorFunc
	for _, caller := range cand.callers {
		switch caller.errorFunc.Type {
		case cand.errorFunc.Type:
			// Same type chain

		case result.WrapperAsType:
			if cand.errorFunc.Type != result.WrapperAs {
				continue
			}
			// AsType calling As

		default:
			// Mismatched wrapper types
			continue
		}

		propagate(wrappers, caller)
	}
}
