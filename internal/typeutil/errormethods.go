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

package typeutil

import "go/types"

var (
	// errorType is the `error` interface type from the universal scope.
	errorType = types.Universe.Lookup("error").Type()

	// anyType is the `any` interface type from the universal scope.
	anyType = types.Universe.Lookup("any").Type()

	// errorSig is the function signature of an `Error() string` method.
	errorSig = sigType{[]types.Type{}, []types.Type{types.Typ[types.String]}}

	// isSig is the function signature of an `Is(error) bool` method.
	isSig = sigType{[]types.Type{errorType}, []types.Type{types.Typ[types.Bool]}}

	// asSig is the function signature of an `As(any) bool` method.
	asSig = sigType{[]types.Type{anyType}, []types.Type{types.Typ[types.Bool]}}

	// unwrapSig is the function signature of an `Unwrap() error` method.
	unwrapSig = sigType{[]types.Type{}, []types.Type{errorType}}

	// unwrapMultipleSig is the function signature of an `Unwrap() []error` method.
	unwrapMultipleSig = sigType{[]types.Type{}, []types.Type{types.NewSlice(errorType)}}
)

// sigType is the parameter and result types of a function signature.
type sigType struct {
	params  []types.Type
	results []types.Type
}

// matchSignature checks if a function signature matches the given parameter and result types.
func (s sigType) matchSignature(sig *types.Signature) bool {
	params, results := sig.Params(), sig.Results()
	if params.Len() != len(s.params) || results.Len() != len(s.results) {
		return false
	}

	for i, param := range s.params {
		if !types.Identical(params.At(i).Type(), param) {
			return false
		}
	}

	for i, result := range s.results {
		if !types.Identical(results.At(i).Type(), result) {
			return false
		}
	}

	return true
}

// HasErrorSig checks whether the provided function signature is `func() string`.
func HasErrorSig(sig *types.Signature) bool {
	return errorSig.matchSignature(sig)
}

// HasIsSig checks whether the provided function signature is `func(error) bool`.
func HasIsSig(sig *types.Signature) bool {
	return isSig.matchSignature(sig)
}

// HasAsSig checks whether the provided function signature is `func(any) bool`.
func HasAsSig(sig *types.Signature) bool {
	return asSig.matchSignature(sig)
}

// HasUnwrapSig checks whether the provided function signature is `func() error` or `func() []error`.
func HasUnwrapSig(sig *types.Signature) bool {
	return unwrapSig.matchSignature(sig) || unwrapMultipleSig.matchSignature(sig)
}
