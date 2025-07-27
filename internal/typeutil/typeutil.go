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
	switch typ := t.(type) {
	case *types.Named:
		return typ.Obj(), false, true

	case *types.Alias:
		return typ.Obj(), false, true

	case *types.Pointer:
		return handlePointer(typ)

	default:
		// Anonymous types (struct literals, nil, etc.)
		return nil, false, false
	}
}

func handlePointer(p *types.Pointer) (tn *types.TypeName, isPtr, ok bool) {
	switch elem := p.Elem().(type) {
	case *types.Named:
		return elem.Obj(), true, true

	case *types.Alias:
		return elem.Obj(), true, true
	}

	// Pointer to anonymous type (e.g., *struct{})
	return nil, true, false
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
	lastTypeExpr := results.List[len(results.List)-1].Type

	// Check if the return type is a type with an `Error() string`` method.
	tv, ok := info.Types[lastTypeExpr]
	if ok && HasErrorMethod(tv.Type) { // inclding concrete types, otherwise: && types.IsInterface(tv.Type)
		return results.NumFields() - 1
	}

	return -1 // Not an error type
}

// HasErrorMethod checks if a given type implements the standard `error`
// interface. Note that when T implements `error`, *T can, but must not, implement `error` too.
func HasErrorMethod(typ types.Type) bool {
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
// Returns true if the signature matches, otherwise false.
func HasErrorSig(sig *types.Signature) bool {
	if sig.Params().Len() > 0 || sig.Results().Len() != 1 {
		return false // Wrong signature
	}

	restype := types.Unalias(sig.Results().At(0).Type())
	if b, basic := restype.(*types.Basic); !basic || b.Kind() != types.String {
		return false // Wrong result type
	}

	return true
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
