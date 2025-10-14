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

// HasErrorMethod checks if a given type implements the standard `error` interface.
// Note that when T implements `error`, *T can, but must not, implement `error` too.
func HasErrorMethod(typ types.Type) bool {
	return HasMethod(typ, "Error", HasErrorSig)
}

// HasMethod checks if a given type implements a method.
// Note that when T implements the method, *T can, but must not, implement the method too.
func HasMethod(typ types.Type, name string, sigCheck func(*types.Signature) bool) bool {
	if typ == UniverseError {
		return true
	}

	obj, _, _ := types.LookupFieldOrMethod(typ, false, nil, name)
	if obj == nil {
		return false // Method not found
	}

	fun, ok := obj.(*types.Func)
	if !ok || !sigCheck(fun.Signature()) {
		return false // *types.Var or wrong signature
	}

	return true
}

// LookupMethod finds a method with the specified name in a type, checking its signature and accounting for embedding.
// Returns the method if found, whether it was found via indirection, and a boolean indicating success.
func LookupMethod(typ types.Type, name string, sigCheck func(*types.Signature) bool) (fun *types.Func, indirect, embedded, found bool) {
	obj, index, indirect := types.LookupFieldOrMethod(typ, true, nil, name)
	if obj == nil {
		return nil, false, false, false // No method with name
	}

	fun, ok := obj.(*types.Func)
	if !ok || !sigCheck(fun.Signature()) {
		return nil, false, false, false // *types.Var or wrong signature
	}

	embedded = len(index) > 1

	return fun, indirect, embedded, true
}
