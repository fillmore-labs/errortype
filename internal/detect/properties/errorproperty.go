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

package properties

import "fillmore-labs.com/errortype/detect/result"

//go:generate go tool bitmask -type ErrorProperty

// ErrorProperty is a bitmask representing properties of an error type's definition
// and usage. These properties are collected to determine whether a type should be
// consistently used as a pointer (*T) or a value (T).
type ErrorProperty uint32

// Properties are grouped by the heuristic that discovers them.
const (
	// --- Properties from user-defined overrides ---.

	// SuppressOverride is set if usage checks for this type are explicitly suppressed.
	SuppressOverride ErrorProperty = 1 << iota // SuppressOverride
	// PointerOverride is set if the type is explicitly marked as a pointer type in overrides.
	PointerOverride // PointerOverride
	// ValueOverride is set if the type is explicitly marked as a value type in overrides.
	ValueOverride // ValueOverride

	// --- Properties from method receivers ---.

	// PointerReceiver is set if the `Error` method has a pointer receiver,
	// e.g., `func (e *MyError) Error() string`. This forces the error to be a pointer type.
	PointerReceiver // PointerReceiver

	// PointerMethod is set if one of the `error` methods `Unwrap`, `Is` or `As`
	// have a pointer receiver, e.g., `func (e *MyError) Unwrap() error`.
	PointerMethod // PointerMethod

	// Embedded is set if the `Error` method is not directly defined on the type,
	// but embedded. This is no indicator and for diagnostics only.
	Embedded // Embedded
	// EmbeddedMethod is set if one of the `error` methods `Is` or `As` is not
	// directly defined on the type, but embedded. This is no indicator and for diagnostics only.
	EmbeddedMethod // EmbeddedMethod

	// --- Properties from type aliases (e.g., `type T = V`) ---.

	// PointerAlias is set for an alias to an imported pointer-type error.
	PointerAlias // PointerAlias
	// ValueAlias is set for an alias to an imported value-type error.
	ValueAlias // ValueAlias

	// --- Properties from variable declarations (e.g., `var ErrSomething = ...`) ---.

	// PointerVar is set for pointer usage, e.g., `var _ error = &T{}` or `var Err = &T{}`.
	PointerVar // PointerVar
	// ValueVar is set for value usage, e.g., `var _ error = T{}` or `var Err = T{}`.
	ValueVar // ValueVar

	// --- Properties from usage in return statements ---.

	// PointerReturn is set for pointer usage, e.g., `return &T{}`.
	PointerReturn // PointerReturn
	// ValueReturn is set for value usage, e.g., `return T{}`.
	ValueReturn // ValueReturn

	// --- Properties from usage in type assertions ---.

	// PointerAssert is set for pointer usage, e.g., `err.(*T)`.
	PointerAssert // PointerAssert
	// ValueAssert is set for value usage, e.g., `err.(T)`.
	ValueAssert // ValueAssert

	// --- Properties from targets in errors.As-like functions ---.

	// PointerTarget is set for pointer usage, e.g., `val target *T; ... errors.As(err, &target)`.
	PointerTarget // PointerTarget
	// ValueTarget is set for value usage, e.g., `val target T; ... errors.As(err, &target)`.
	ValueTarget // ValueTarget

	// --- Properties from usage in composite literals ---.

	// PointerLiteral is set for pointer usage, e.g., `&T{}`.
	PointerLiteral // PointerLiteral
	// ValueLiteral is set for value usage, e.g., `T{}`.
	ValueLiteral // ValueLiteral

	// --- Properties from usage in type casts ---.

	// PointerCast is set for pointer usage, e.g., `(*T)(v)`.
	PointerCast // PointerCast
	// ValueCast is set for value usage, e.g., `(T)(v)`.
	ValueCast // ValueCast

	// --- Properties from other method receivers ---.

	// PointerReceivers is set if all methods on the type have pointer receivers (a weak indicator).
	PointerReceivers // PointerReceivers
	// ValueReceivers is set if all methods on the type have value receivers (a weak indicator).
	ValueReceivers // ValueReceivers

	// --- Properties from type definition ---.

	// PointerDef is set for defined pointer types, e.g., `type T *S`. Such types are used like values.
	PointerDef // PointerDef

	// NonStruct is set if the defined type is not a structure.
	NonStruct // NonStruct

	// ZeroSized is set if the defined type is a zero-sized structure or pointer.
	ZeroSized // ZeroSized

	// --- Others ---.

	// OverrideMask identifies any override property.
	OverrideMask = PointerOverride | ValueOverride | SuppressOverride

	// ReceiverMask identifies error method receivers.
	ReceiverMask = PointerReceiver | PointerMethod
)

// _propertyPairs defines the categories of evidence used to determine if an error
// type is a pointer or value type. The pairs are ordered by precedence, from
// the strongest evidence to the weakest. The first category with a non-contradictory
// signal determines the type.
var _propertyPairs = [...]struct{ pointerProp, valueProp ErrorProperty }{
	{PointerAlias, ValueAlias},         // Aliases of imported error types.
	{PointerVar, ValueVar},             // Sentinel errors or `var _ error` assertions.
	{PointerReturn, ValueReturn},       // Usage in `return` statements.
	{PointerAssert, ValueAssert},       // Usage in type assertions.
	{PointerTarget, ValueTarget},       // Usage in errors.As-like functions.
	{PointerLiteral, ValueLiteral},     // Usage as a composite literal.
	{PointerCast, ValueCast},           // Usage in type casts.
	{PointerReceivers, ValueReceivers}, // Weakest: consistency of other method receivers.
}

// DeterminedType checks if the collected properties unambiguously determine
// whether the type should be a pointer or a value error type.
//
// Contradictory properties (e.g., both PointerVar and ValueVar being set)
// for a given category are ignored, and the decision moves to the next category.
func (e ErrorProperty) DeterminedType() result.ErrorType {
	switch e & OverrideMask { // Overrides have the highest precedence.
	case SuppressOverride:
		return result.Suppress

	case PointerOverride:
		return result.Pointer

	case ValueOverride:
		return result.Value
	}

	// Errors with pointer receivers can only be used in only one way.
	// Errors with an `Unwrap() error` method with pointer receiver would behave differently as values.
	if e&ReceiverMask != 0 {
		return result.Pointer
	}

	for _, pair := range _propertyPairs {
		switch pointerProp, valueProp := pair.pointerProp, pair.valueProp; e & (pointerProp | valueProp) { // Check for a non-contradictory usage within this category.
		case pointerProp:
			return result.Pointer

		case valueProp:
			return result.Value
		}
	}

	// A special case for defined pointer types like `type T *S`.
	// Although the underlying type is a pointer, `T` itself is used as a value
	// (e.g., you return `T`, not `*T`), so we treat it as a value type.
	if e&PointerDef != 0 {
		return result.Value
	}

	// No unambiguous usage was found.
	return result.Undecided
}
