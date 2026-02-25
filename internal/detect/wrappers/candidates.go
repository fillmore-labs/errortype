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

	"fillmore-labs.com/errortype/detect/result"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// wrapperCandidate represents a function with a signature matching a potential wrapper, tracking its context to build a call graph.
type wrapperCandidate struct {
	fun       *types.Func
	body      *ast.BlockStmt
	srcVar    *types.Var
	tgtVar    *types.Var
	tParam    *types.TypeParam
	callers   []*wrapperCandidate
	errorFunc result.ErrorFunc
}

// findCandidates scans the package-level declarations to discover functions whose signatures make them potential wrapper candidates.
// Functions already classified as wrappers (seeds or explicit overrides in known) are skipped, so their classification stays authoritative.
func findCandidates(info *types.Info, files []*ast.File, known result.ErrorFuncs) map[types.Object]*wrapperCandidate {
	var wrapperCandidates map[types.Object]*wrapperCandidate

	for decl := range typeutil.AllFuncDecls(files) {
		if decl.Body == nil {
			continue
		}

		fun, ok := info.Defs[decl.Name].(*types.Func)
		if !ok { // should not happen
			continue
		}

		if _, ok := known[fun]; ok {
			continue // already classified as a seed or override; do not re-derive
		}

		if wrapper := candidateWrapper(fun, decl.Body); wrapper != nil {
			if wrapperCandidates == nil {
				wrapperCandidates = make(map[types.Object]*wrapperCandidate)
			}
			wrapperCandidates[fun] = wrapper
		}
	}

	return wrapperCandidates
}

func candidateWrapper(fun *types.Func, body *ast.BlockStmt) *wrapperCandidate {
	sig := fun.Signature()
	params := sig.Params()
	nParams := min(params.Len(), maxParamIndex)

	srcIdx := firstErrorParameter(params, nParams)

	if srcIdx < 0 {
		// Only check for Errorf wrappers when no error parameters are found
		return errorfCandidate(fun, body, sig, params)
	}

	srcParam := params.At(srcIdx)
	if srcIdx < nParams-1 {
		tgtIdx := srcIdx + 1
		switch tgtParam := params.At(tgtIdx); {
		case typeutil.IsErrorInterface(tgtParam.Type()):
			return candidate(fun, body, result.WrapperIs, srcParam, tgtParam, srcIdx, tgtIdx)

		case typeutil.IsAnyInterface(tgtParam.Type()):
			return candidate(fun, body, result.WrapperAs, srcParam, tgtParam, srcIdx, tgtIdx)
		}

		// Not a wrapper of type Is or As, check for AsType
	}

	tParams := sig.TypeParams()
	nTParams := min(tParams.Len(), maxParamIndex)

	if tgtIdx := firstErrorTypeParameter(tParams, nTParams); tgtIdx >= 0 {
		tParam := tParams.At(tgtIdx)

		return candidateT(fun, body, result.WrapperAsType, srcParam, tParam, srcIdx, tgtIdx)
	}

	return nil
}

func errorfCandidate(fun *types.Func, body *ast.BlockStmt, sig *types.Signature, params *types.Tuple) *wrapperCandidate {
	if !sig.Variadic() {
		return nil
	}

	srcIdx := params.Len() - 2
	if srcIdx < 0 || srcIdx >= maxParamIndex {
		return nil
	}

	srcParam := params.At(srcIdx)
	if srcParam.Type() != types.Typ[types.String] {
		return nil
	}

	tgtIdx := srcIdx + 1
	tgtParam := params.At(tgtIdx)

	if slice, ok := tgtParam.Type().(*types.Slice); !ok || !typeutil.IsAnyInterface(slice.Elem()) {
		return nil
	}

	return candidate(fun, body, result.WrapperErrorf, srcParam, tgtParam, srcIdx, tgtIdx)
}

func candidate(fun *types.Func, body *ast.BlockStmt, typ result.WrapperType, srcParam, tgtParam *types.Var, srcIdx, tgtIdx int) *wrapperCandidate {
	return &wrapperCandidate{
		fun:       fun,
		errorFunc: result.ErrorFunc{Type: typ, Source: int8(srcIdx), Target: int8(tgtIdx)}, // #nosec G115 -- limited by nParams.
		body:      body,
		srcVar:    srcParam,
		tgtVar:    tgtParam,
	}
}

func candidateT(fun *types.Func, body *ast.BlockStmt, typ result.WrapperType, srcParam *types.Var, tParam *types.TypeParam, srcIdx, tgtIdx int) *wrapperCandidate {
	return &wrapperCandidate{
		fun:       fun,
		errorFunc: result.ErrorFunc{Type: typ, Source: int8(srcIdx), Target: int8(tgtIdx)}, // #nosec G115 -- limited by nParams, nTParams.
		body:      body,
		srcVar:    srcParam,
		tParam:    tParam,
	}
}
