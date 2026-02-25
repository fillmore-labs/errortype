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
	"go/types"
	"math"

	"fillmore-labs.com/errortype/detect/result"
	"fillmore-labs.com/errortype/internal/typeutil"
)

const maxParamIndex = math.MaxInt8 - 1 // reserves room for srcIdx+1 in int8

func findParameters(sig *types.Signature, wrapperType result.WrapperType) (result.ErrorFunc, bool) {
	params := sig.Params()

	switch nParams := min(params.Len(), maxParamIndex); wrapperType {
	case result.WrapperIs:
		srcIdx := firstErrorParameter(params, nParams-1)
		if srcIdx < 0 || !typeutil.IsErrorInterface(params.At(srcIdx+1).Type()) {
			break
		}

		return result.ErrorFunc{Type: wrapperType, Source: int8(srcIdx), Target: int8(srcIdx + 1)}, true // #nosec G115 -- limited by nParams.

	case result.WrapperAs:
		srcIdx := firstErrorParameter(params, nParams-1)
		if srcIdx < 0 || !typeutil.IsAnyInterface(params.At(srcIdx+1).Type()) {
			break
		}

		return result.ErrorFunc{Type: wrapperType, Source: int8(srcIdx), Target: int8(srcIdx + 1)}, true // #nosec G115 -- limited by nParams.

	case result.WrapperAsType:
		srcIdx := firstErrorParameter(params, nParams)
		if srcIdx < 0 {
			break
		}

		tParams := sig.TypeParams()
		nTParams := min(tParams.Len(), maxParamIndex)

		tgtIdx := firstErrorTypeParameter(tParams, nTParams)
		if tgtIdx < 0 {
			break
		}

		return result.ErrorFunc{Type: wrapperType, Source: int8(srcIdx), Target: int8(tgtIdx)}, true // #nosec G115 -- limited by nParams, nTParams.

	case result.WrapperErrorf:
		if !sig.Variadic() {
			break
		}

		srcIdx := params.Len() - 2
		if srcIdx < 0 || srcIdx >= maxParamIndex || params.At(srcIdx).Type() != types.Typ[types.String] {
			break
		}

		return result.ErrorFunc{Type: wrapperType, Source: int8(srcIdx), Target: int8(srcIdx + 1)}, true

	case result.FuncIsType, result.FuncEqual:
		srcIdx := firstAnyParameter(params, nParams-1)
		if srcIdx < 0 || !typeutil.IsAnyInterface(params.At(srcIdx+1).Type()) {
			break
		}

		return result.ErrorFunc{Type: wrapperType, Source: int8(srcIdx), Target: int8(srcIdx + 1)}, true // #nosec G115 -- limited by nParams.

	case result.FuncAssert:
		tParams := sig.TypeParams()
		nTParams := min(tParams.Len(), maxParamIndex)

		tgtIdx := firstAnyTypeParameter(tParams, nTParams)
		if tgtIdx < 0 {
			break
		}

		return result.ErrorFunc{
			Type:   wrapperType,
			Source: -1,
			Target: int8(tgtIdx), // #nosec G115 -- limited by nTParams.
		}, true
	}

	return result.ErrorFunc{}, false
}

func firstErrorParameter(params *types.Tuple, nParams int) int {
	for idx := range nParams {
		if typeutil.IsErrorInterface(params.At(idx).Type()) {
			return idx
		}
	}

	return -1
}

func firstAnyParameter(params *types.Tuple, nParams int) int {
	for idx := range nParams {
		if typeutil.IsAnyInterface(params.At(idx).Type()) {
			return idx
		}
	}

	return -1
}

func firstErrorTypeParameter(tParams *types.TypeParamList, nTParams int) int {
	for idx := range nTParams {
		if typeutil.HasErrorMethod(tParams.At(idx).Constraint()) {
			return idx
		}
	}

	return -1
}

func firstAnyTypeParameter(tParams *types.TypeParamList, nTParams int) int {
	for idx := range nTParams {
		if typeutil.IsAnyInterface(tParams.At(idx).Constraint()) {
			return idx
		}
	}

	return -1
}
