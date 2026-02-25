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

import (
	"go/ast"
	"go/types"
)

const (
	// errorName is the name of the `Error() string` method.
	errorName = "Error"

	// isName is the name of the `Is(err error) bool` method.
	isName = "Is"

	// asName is the name of the `As(err any) bool` method.
	asName = "As"

	// unwrapName is the name of the `Unwrap() error` and `Unwrap() []error` methods.
	unwrapName = "Unwrap"
)

// SignatureCheckFor returns a signature validation function for a method.
// It supports error tree-related methods `Is`, `As`, and `Unwrap`.
// If the method name is not one of these, it returns nil.
func SignatureCheckFor(name string) func(*types.Signature) bool {
	switch name {
	case isName:
		return isSig.matchSig

	case unwrapName:
		return matchUnwrapSig

	case asName:
		return asSig.matchSig
	}

	return nil
}

// IsIsMethod cehcks for an `Is` method declaration with body.
func IsIsMethod(f *ast.FuncDecl) bool {
	return f.Name.Name == isName && f.Body != nil
}

var (
	// ErrorMethod represents the standard `Error() string` method required by the built-in error interface.
	ErrorMethod = Method{
		matchSig: errorSig.matchSig,
		name:     errorName,
	}
	// IsMethod represents the `Is(error) bool` method used to customize [errors.Is] behavior.
	IsMethod = Method{
		matchSig: isSig.matchSig,
		name:     isName,
	}
	// AsMethod represents the `As(any) bool` method used to customize [errors.As] behavior.
	AsMethod = Method{
		matchSig: asSig.matchSig,
		name:     asName,
	}
	// UnwrapMethod represents the `Unwrap() error` or `Unwrap() []error` method used to retrieve a wrapped error.
	UnwrapMethod = Method{
		matchSig: matchUnwrapSig,
		name:     unwrapName,
	}
	// UnwrapMultipleMethod represents the `Unwrap() []error` method used by multi-error wrappers like [errors.Join].
	UnwrapMultipleMethod = Method{
		matchSig: unwrapMultipleSig.matchSig,
		name:     unwrapName,
	}
)

// sigType is the parameter and result types of a function signature.
type sigType struct {
	params  []types.Type
	results []types.Type
}

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

// matchSig checks if a function signature matches the given parameter and result types.
func (s sigType) matchSig(sig *types.Signature) bool {
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

// matchUnwrapSig checks for `Unwrap() error` or `Unwrap() []error`.
func matchUnwrapSig(sig *types.Signature) bool {
	return unwrapSig.matchSig(sig) || unwrapMultipleSig.matchSig(sig)
}
