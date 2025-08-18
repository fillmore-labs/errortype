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
	"go/token"
	"go/types"
)

// errorIsInterface holds a reference to the `interface{ Is(error) bool }` type.
// This is used by [shouldSuppressDiagnostic] to check if a type implements
// the optional error comparison interface defined by `errors.Is`.
//
//nolint:gochecknoglobals
var (
	universeError = types.Universe.Lookup("error").Type()

	// errorIsInterface represents `interface{ Is(error) bool }`.
	errorIsInterface = newErrorIsInterface(universeError)

	// errorUnwrapInterface represents `interface{ Unwrap() error }`.
	errorUnwrapInterface = newErrorUnwrapInterface(universeError)

	// errorUnwrapArrayInterface represents `interface{ Unwrap() []error }`.
	errorUnwrapArrayInterface = newErrorUnwrapArrayInterface(universeError)
)

// newErrorIsInterface constructs and returns a new [types.Interface] representing
// the `interface{ Is(error) bool }` type.
func newErrorIsInterface(universeError types.Type) *types.Interface {
	const isMethodName = "Is"

	var noPkg *types.Package

	params := singleVar(universeError)
	results := singleVar(types.Typ[types.Bool])
	sig := types.NewSignatureType(nil, nil, nil, params, results, false)
	isFunc := types.NewFunc(token.NoPos, noPkg, isMethodName, sig)

	return interfaceOf(isFunc)
}

// newErrorUnwrapInterface constructs and returns a new [types.Interface] representing
// the `interface{ Unwrap() error }` type.
func newErrorUnwrapInterface(universeError types.Type) *types.Interface {
	const unwrapMethodName = "Unwrap"

	var noPkg *types.Package

	results := singleVar(universeError)
	sig := types.NewSignatureType(nil, nil, nil, nil, results, false)
	unwrapFunc := types.NewFunc(token.NoPos, noPkg, unwrapMethodName, sig)

	return interfaceOf(unwrapFunc)
}

// newErrorUnwrapArrayInterface constructs and returns a new [types.Interface] representing
// the `interface{ Unwrap() []error }` type.
func newErrorUnwrapArrayInterface(universeError types.Type) *types.Interface {
	const unwrapMethodName = "Unwrap"

	var noPkg *types.Package

	results := singleVar(types.NewSlice(universeError))
	sig := types.NewSignatureType(nil, nil, nil, nil, results, false)
	unwrapFunc := types.NewFunc(token.NoPos, noPkg, unwrapMethodName, sig)

	return interfaceOf(unwrapFunc)
}

// singleVar constructs a [types.Tuple] containing a single unnamed variable
// of the specified [types.Type].
func singleVar(t types.Type) *types.Tuple {
	const noName = ""

	var noPkg *types.Package

	return types.NewTuple(types.NewVar(token.NoPos, noPkg, noName, t))
}

func interfaceOf(method *types.Func) *types.Interface {
	return types.NewInterfaceType([]*types.Func{method}, nil).Complete()
}
