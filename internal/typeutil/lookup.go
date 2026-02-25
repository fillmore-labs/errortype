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
	"go/types"

	"golang.org/x/tools/go/types/typeutil"
)

// IsErrorInterface checks if the provided type is the built-in `error` interface.
func IsErrorInterface(typ types.Type) bool {
	return typ == errorType
}

// IsAnyInterface determines if the given type is the `any` interface or an interface with no methods.
func IsAnyInterface(typ types.Type) bool {
	if typ == anyType {
		return true
	}

	iface, ok := typ.Underlying().(*types.Interface)

	return ok && iface.NumMethods() == 0
}

// IsInterfaceWithError checks if the provided type is an error-like interface.
func IsInterfaceWithError(typ types.Type) bool {
	if IsErrorInterface(typ) {
		return true
	}

	iface, ok := typ.Underlying().(*types.Interface)

	return ok && ErrorMethod.HasMethod(iface, false)
}

// HasErrorMethod checks if a given type directly implements the standard `error` interface.
// Note that when a pointer receiver *T implements `error`, the value type T does not necessarily implement it.
func HasErrorMethod(typ types.Type) bool {
	if IsErrorInterface(typ) {
		return true
	}

	return ErrorMethod.HasMethod(typ, false)
}

// HasErrorMethodCached checks if a given type directly implements the standard `error` interface.
func HasErrorMethodCached(cache *typeutil.Map, typ types.Type) bool {
	if IsErrorInterface(typ) {
		return true
	}

	if v, ok := cache.At(typ).(bool); ok {
		return v
	}

	v := ErrorMethod.HasMethod(typ, false)
	cache.Set(typ, v)

	return v
}

// Method represents a specific error-related method type, defining its name and
// signature validation function.
type Method struct {
	matchSig func(*types.Signature) bool
	name     string
}

// MatchSig matches the signature of sig.
func (m Method) MatchSig(sig *types.Signature) bool {
	return m.matchSig(sig)
}

// HasMethod checks if a given type implements a method.
//
// If addressable is set, typ is the type of an addressable variable.
func (m Method) HasMethod(typ types.Type, addressable bool) bool {
	fun, _, _ := lookupMethod(typ, addressable, m.name)

	return fun != nil && m.matchSig(fun.Signature())
}

// LookupResult is the result of a [Method.Lookup] on a type.
type LookupResult struct {
	Recv     *types.Var // Recv is the receiver variable of the resolved method.
	Indirect bool       // Indirect indicates if the method was looked up via pointer indirection.
	Embedded bool       // Embedded indicates if the method is promoted from an embedded field.
}

// Lookup finds a method with the specified name in a type, checking its signature and accounting for embedding.
// Returns the method if found, whether it was found via indirection, and a boolean indicating success.
func (m Method) Lookup(typ types.Type) (res LookupResult, ok bool) {
	fun, index, indirect := lookupMethod(typ, true, m.name)
	if fun == nil || !m.matchSig(fun.Signature()) {
		return res, false // *types.Var or wrong signature
	}

	res.Recv = fun.Signature().Recv()
	res.Indirect = indirect
	res.Embedded = len(index) > 1

	return res, true
}

func lookupMethod(typ types.Type, addressable bool, name string) (fun *types.Func, index []int, indirect bool) {
	obj, index, indirect := types.LookupFieldOrMethod(typ, addressable, nil, name)
	fun, _ = obj.(*types.Func)

	return fun, index, indirect
}
