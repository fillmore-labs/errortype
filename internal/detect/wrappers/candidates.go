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

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/detect/result"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// wrapperCandidate represents a function with a signature matching a potential wrapper, tracking its context to build a call graph.
type wrapperCandidate struct {
	fun       *types.Func
	errorFunc result.ErrorFunc
	body      *ast.BlockStmt
	srcVar    *types.Var
	tgtVar    *types.Var       // non-nil for Is/As candidates
	tParam    *types.TypeParam // non-nil for AsType candidates
	callers   []*wrapperCandidate
}

// findCandidates scans the package-level declarations to discover functions whose signatures make them potential wrapper candidates.
func findCandidates(p *analysis.Pass) map[*types.Func]*wrapperCandidate {
	var wrapperCandidates map[*types.Func]*wrapperCandidate

	for decl := range typeutil.AllFuncDecls(p.Files) {
		if decl.Body == nil {
			continue
		}

		fun, ok := p.TypesInfo.Defs[decl.Name].(*types.Func)
		if !ok { // should not happen
			continue
		}

		if cand := candidateWrapper(fun, decl.Body); cand != nil {
			if wrapperCandidates == nil {
				wrapperCandidates = make(map[*types.Func]*wrapperCandidate)
			}
			wrapperCandidates[fun] = cand
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
		return nil
	}

	srcParam := params.At(srcIdx)
	if srcIdx < nParams-1 {
		switch tgtParam := params.At(srcIdx + 1); {
		case typeutil.IsErrorInterface(tgtParam.Type()):
			return &wrapperCandidate{
				fun:       fun,
				errorFunc: result.ErrorFunc{Type: result.WrapperIs, Source: int8(srcIdx), Target: int8(srcIdx + 1)}, // #nosec:G115 -- limited by nParams
				body:      body,
				srcVar:    srcParam,
				tgtVar:    tgtParam,
			}

		case typeutil.IsAnyInterface(tgtParam.Type()):
			return &wrapperCandidate{
				fun:       fun,
				errorFunc: result.ErrorFunc{Type: result.WrapperAs, Source: int8(srcIdx), Target: int8(srcIdx + 1)}, // #nosec:G115 -- limited by nParams
				body:      body,
				srcVar:    srcParam,
				tgtVar:    tgtParam,
			}
		}

		// Not a wrapper of type Is or As, check for AsType
	}

	tParams := sig.TypeParams()
	nTParams := min(tParams.Len(), maxParamIndex)

	if tgtIdx := firstErrorTypeParameter(tParams, nTParams); tgtIdx >= 0 {
		tParam := tParams.At(tgtIdx)

		return &wrapperCandidate{
			fun:       fun,
			errorFunc: result.ErrorFunc{Type: result.WrapperAsType, Source: int8(srcIdx), Target: int8(tgtIdx)}, // #nosec:G115 -- limited by nParams, nTParams
			body:      body,
			srcVar:    srcParam,
			tParam:    tParam,
		}
	}

	return nil
}
