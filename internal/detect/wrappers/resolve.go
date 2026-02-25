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

// resolveCandidates traverses the AST of wrapper candidates to find internal calls to known wrappers or other candidates.
func resolveCandidates(info *types.Info, wrapperCandidates map[types.Object]*wrapperCandidate, known result.ErrorFuncs) result.ErrorFuncs {
	wrappers := make(result.ErrorFuncs)

	for _, cand := range wrapperCandidates {
		// We do not use [golang.org/x/tools/go/ast/inspector.Inspector] here since we assume
		// most packages don't contain wrappers, and when, they are few and small.
	bodyanalysis:
		for n := range ast.Preorder(cand.body) {
			switch n := n.(type) {
			case *ast.CallExpr:
				if matchCall(info, cand, n, wrapperCandidates, known, wrappers) {
					propagate(wrappers, cand)
					break bodyanalysis // first matching call only
				}

			case *ast.AssignStmt:
				for _, lhs := range n.Lhs {
					if matchArgs(info, lhs, cand.srcVar, cand.tgtVar) {
						break bodyanalysis // Argument is modified
					}
				}

			case *ast.UnaryExpr:
				if n.Op == token.AND {
					if matchArgs(info, n.X, cand.srcVar, cand.tgtVar) {
						break bodyanalysis // Argument has address taken
					}
				}
			}
		}
	}

	return wrappers
}

func matchCall(info *types.Info, cand *wrapperCandidate, call *ast.CallExpr, wrapperCandidates map[types.Object]*wrapperCandidate, known, wrappers result.ErrorFuncs) bool {
	fun, ok := typeutil.FuncOf(info, call)
	if !ok {
		return false // Could not resolve the function, might be a func variable.
	}

	// Is it a call to an already known wrapper?
	ef, ok := known[fun.Func]
	if !ok {
		ef, ok = wrappers[fun.Func]
	}

	if ok {
		return checkKnownWrapper(info, fun, call.Args, ef, cand)
	}

	// Is it a call to another wrapperCandidate wrapper?
	if callee, ok := wrapperCandidates[fun.Func]; ok {
		registerCaller(info, fun, call.Args, callee, cand)
	}

	return false
}

func checkKnownWrapper(info *types.Info, fun typeutil.ResolvedFunc, args []ast.Expr, ef result.ErrorFunc, cand *wrapperCandidate) bool {
	switch typ := cand.errorFunc.Type; typ {
	case result.WrapperIs, result.WrapperAs, result.WrapperErrorf:
		return matchWrapperArgs(info, fun, args, ef, typ, cand.srcVar, cand.tgtVar)

	case result.WrapperAsType:
		return matchWrapperType(info, fun, args, ef, cand.srcVar, cand.tParam)
	}

	return false
}

func registerCaller(info *types.Info, fun typeutil.ResolvedFunc, args []ast.Expr, callee, cand *wrapperCandidate) {
	switch typ := cand.errorFunc.Type; typ {
	case result.WrapperIs, result.WrapperAs, result.WrapperErrorf:
		if matchWrapperArgs(info, fun, args, callee.errorFunc, typ, cand.srcVar, cand.tgtVar) {
			callee.callers = append(callee.callers, cand)
		}

	case result.WrapperAsType:
		if matchWrapperType(info, fun, args, callee.errorFunc, cand.srcVar, cand.tParam) {
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
			continue // Mismatched wrapper types
		}

		propagate(wrappers, caller)
	}
}
