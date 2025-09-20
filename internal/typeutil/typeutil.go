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

// TypeNameOf extracts the underlying type name from a given type.
// It handles pointers, dereferencing them to find the core [types.TypeName].
//
// It returns the found [types.TypeName] type, a boolean indicating if the original type
// was a pointer, and a boolean indicating if a type name was successfully found.
// It returns false for anonymous types (like struct literals).
func TypeNameOf(t types.Type) (tn *types.TypeName, isPtr, ok bool) {
	isPtr = false

	for {
		switch typ := t.(type) {
		case *types.Named:
			return typ.Obj(), isPtr, true

		case *types.Alias:
			return typ.Obj(), isPtr, true

		case *types.Pointer:
			if isPtr {
				// Double Pointer
				return nil, isPtr, false
			}

			t = typ.Elem()
			isPtr = true

			continue

		default:
			// Anonymous types (struct literals, nil, etc.)
			// We are also not interested in type parameters or basic types
			return nil, isPtr, false
		}
	}
}

// HasErrorResult checks whether the given function result list has an error type as its last return value.
// Returns the index of the error result or -1 when not found.
func HasErrorResult(info *types.Info, results *ast.FieldList) int {
	// We are only interested in functions with return values.
	if results == nil || len(results.List) == 0 {
		return -1 // No result
	}

	// Only check the last return type expression, as `error` is
	// conventionally the last one.
	lastType := results.List[len(results.List)-1].Type

	// Check if the return type is a type with an `Error() string` method.
	if tv, ok := info.Types[lastType]; ok && types.IsInterface(tv.Type) && HasErrorMethod(tv.Type) {
		return results.NumFields() - 1
	}

	return -1 // Not an error type
}

// HasErrorMethod checks if a given type implements the standard `error` interface.
// Note that when T implements `error`, *T can, but must not, implement `error` too.
func HasErrorMethod(typ types.Type) bool {
	if typ == UniverseError {
		return true
	}

	obj, _, _ := types.LookupFieldOrMethod(typ, false, nil, "Error")
	if obj == nil {
		return false // Not an error type
	}

	fun, ok := obj.(*types.Func)
	if !ok || !HasErrorSig(fun.Signature()) {
		return false // *types.Var or wrong signature
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

// HasUnwrapSig checks whether the provided function signature is `func() error`.
func HasUnwrapSig(sig *types.Signature) bool {
	return unwrapSig.matchSignature(sig) || unwrapMultipleSig.matchSignature(sig)
}

// HasPointerReceiver determines whether the given method signature has a pointer receiver.
// It returns true if the receiver is a pointer type, and false otherwise.
func HasPointerReceiver(sig *types.Signature) (elem types.Type, isPtr bool) {
	recv := sig.Recv()
	if recv == nil {
		return nil, false // Not a method
	}

	if p, ok := types.Unalias(recv.Type()).(*types.Pointer); ok {
		return p.Elem(), true
	}

	return nil, false
}
