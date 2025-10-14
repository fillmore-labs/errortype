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

import (
	"fillmore-labs.com/errortype/internal/bitflag"
	"fillmore-labs.com/errortype/internal/errortypes"
)

// ErrorProperty is a bitmask representing properties of an error type's definition
// and usage. These properties are collected to determine whether a type should be
// consistently used as a pointer (*T) or a value (T).
type ErrorProperty uint32

// Properties are grouped by the heuristic that discovers them.
const (
	// --- Properties from user-defined overrides ---.

	// SuppressOverride is set if usage checks for this type are explicitly suppressed.
	SuppressOverride ErrorProperty = 1 << posSuppressOverride
	// PointerOverride is set if the type is explicitly marked as a pointer type in overrides.
	PointerOverride ErrorProperty = 1 << posPointerOverride
	// ValueOverride is set if the type is explicitly marked as a value type in overrides.
	ValueOverride ErrorProperty = 1 << posValueOverride

	// --- Properties from method receivers ---.

	// PointerReceiver is set if the `Error` method has a pointer receiver,
	// e.g., `func (e *MyError) Error() string`. This forces the error to be a pointer type.
	PointerReceiver ErrorProperty = 1 << posPointerReceiver

	// PointerMethod is set if one of the `error` methods `Unwrap`, `Is` or `As`
	// have a pointer receiver, e.g., `func (e *MyError) Unwrap() error`.
	PointerMethod ErrorProperty = 1 << posPointerMethod

	// Embedded is set if the `Error` method is not directly defined on the type,
	// but embedded. This is no indicator and for diagnostics only.
	Embedded ErrorProperty = 1 << posEmbedded
	// EmbeddedMethod is set if one of the `error` methods `Is` or `As` is not
	// directly defined on the type, but embedded. This is no indicator and for diagnostics only.
	EmbeddedMethod ErrorProperty = 1 << posEmbeddedMethod

	// --- Properties from type aliases (e.g., `type T = V`) ---.

	// PointerAlias is set for an alias to an imported pointer-type error.
	PointerAlias ErrorProperty = 1 << posPointerAlias
	// ValueAlias is set for an alias to an imported value-type error.
	ValueAlias ErrorProperty = 1 << posValueAlias

	// --- Properties from variable declarations (e.g., `var ErrSomething = ...`) ---.

	// PointerVar is set for pointer usage, e.g., `var _ error = &T{}` or `var Err = &T{}`.
	PointerVar ErrorProperty = 1 << posPointerVar
	// ValueVar is set for value usage, e.g., `var _ error = T{}` or `var Err = T{}`.
	ValueVar ErrorProperty = 1 << posValueVar

	// --- Properties from usage in return statements ---.

	// PointerReturn is set for pointer usage, e.g., `return &T{}`.
	PointerReturn ErrorProperty = 1 << posPointerReturn
	// ValueReturn is set for value usage, e.g., `return T{}`.
	ValueReturn ErrorProperty = 1 << posValueReturn

	// --- Properties from usage in type assertions ---.

	// PointerAssert is set for pointer usage, e.g., `err.(*T)`.
	PointerAssert ErrorProperty = 1 << posPointerAssert
	// ValueAssert is set for value usage, e.g., `err.(T)`.
	ValueAssert ErrorProperty = 1 << posValueAssert

	// --- Properties from targets in errors.As-like functions ---.

	// PointerTarget is set for pointer usage, e.g., `val target *T; ... errors.As(err, &target)`.
	PointerTarget ErrorProperty = 1 << posPointerTarget
	// ValueTarget is set for value usage, e.g., `val target T; ... errors.As(err, &target)`.
	ValueTarget ErrorProperty = 1 << posValueTarget

	// --- Properties from usage in composite literals ---.

	// PointerLiteral is set for pointer usage, e.g., `&T{}`.
	PointerLiteral ErrorProperty = 1 << posPointerLiteral
	// ValueLiteral is set for value usage, e.g., `T{}`.
	ValueLiteral ErrorProperty = 1 << posValueLiteral

	// --- Properties from usage in type casts ---.

	// PointerCast is set for pointer usage, e.g., `(*T)(v)`.
	PointerCast ErrorProperty = 1 << posPointerCast
	// ValueCast is set for value usage, e.g., `(T)(v)`.
	ValueCast ErrorProperty = 1 << posValueCast

	// --- Properties from other method receivers ---.

	// PointerReceivers is set if all methods on the type have pointer receivers (a weak indicator).
	PointerReceivers ErrorProperty = 1 << posPointerReceivers
	// ValueReceivers is set if all methods on the type have value receivers (a weak indicator).
	ValueReceivers ErrorProperty = 1 << posValueReceivers

	// --- Properties from type definition ---.

	// PointerDef is set for defined pointer types, e.g., `type T *S`. Such types are used like values.
	PointerDef ErrorProperty = 1 << posPointerDef

	// NonStruct is set if the defined type is not a structure.
	NonStruct ErrorProperty = 1 << posNonStruct

	// ZeroSized is set if the defined type is a zero-sized structure or pointer.
	ZeroSized ErrorProperty = 1 << posZeroSized

	// --- Others ---.

	// OverrideMask is a bitmask to identify any override property.
	OverrideMask = PointerOverride | ValueOverride | SuppressOverride

	// ReceiverMask is a bitmask to identify error method receivers.
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
// It returns true for `ok` if the pointer-ness is determined.
//
// Contradictory properties (e.g., both PointerVar and ValueVar being set)
// for a given category are ignored, and the decision moves to the next category.
func (e ErrorProperty) DeterminedType() errortypes.ErrorType {
	switch e & OverrideMask { // Overrides have the highest precedence.
	case SuppressOverride:
		return errortypes.SuppressType

	case PointerOverride:
		return errortypes.PointerType

	case ValueOverride:
		return errortypes.ValueType
	}
	// Errors with pointer receivers can only be used in only one way.
	// Errors with an `Unwrap() error` method with pointer receiver would behave differently as values.
	if e&ReceiverMask != 0 {
		return errortypes.PointerType
	}

	for _, pair := range _propertyPairs {
		switch pointerProp, valueProp := pair.pointerProp, pair.valueProp; e & (pointerProp | valueProp) { // Check for a non-contradictory usage within this category.
		case pointerProp:
			return errortypes.PointerType

		case valueProp:
			return errortypes.ValueType
		}
	}

	// A special case for defined pointer types like `type T *S`.
	// Although the underlying type is a pointer, `T` itself is used as a value
	// (e.g., you return `T`, not `*T`), so we treat it as a value type.
	if e&PointerDef != 0 {
		return errortypes.ValueType
	}

	// No unambiguous usage was found.
	return errortypes.UndecidedType
}

const (
	posSuppressOverride = 25 - iota
	posPointerOverride
	posValueOverride
	posPointerReceiver
	posPointerMethod
	posEmbedded
	posEmbeddedMethod
	posPointerAlias
	posValueAlias
	posPointerVar
	posValueVar
	posPointerReturn
	posValueReturn
	posPointerAssert
	posValueAssert
	posPointerTarget
	posValueTarget
	posPointerLiteral
	posValueLiteral
	posPointerCast
	posValueCast
	posPointerReceivers
	posValueReceivers
	posPointerDef
	posNonStruct
	posZeroSized
)

var _errorPropertyNames = [...]string{
	posSuppressOverride: "SuppressOverride",
	posPointerOverride:  "PointerOverride",
	posValueOverride:    "ValueOverride",
	posPointerReceiver:  "PointerReceiver",
	posPointerMethod:    "PointerMethod",
	posEmbedded:         "Embedded",
	posEmbeddedMethod:   "EmbeddedMethod",
	posPointerAlias:     "PointerAlias",
	posValueAlias:       "ValueAlias",
	posPointerVar:       "PointerVar",
	posValueVar:         "ValueVar",
	posPointerReturn:    "PointerReturn",
	posValueReturn:      "ValueReturn",
	posPointerAssert:    "PointerAssert",
	posValueAssert:      "ValueAssert",
	posPointerTarget:    "PointerTarget",
	posValueTarget:      "ValueTarget",
	posPointerLiteral:   "PointerLiteral",
	posValueLiteral:     "ValueLiteral",
	posPointerCast:      "PointerCast",
	posValueCast:        "ValueCast",
	posPointerReceivers: "PointerReceivers",
	posValueReceivers:   "ValueReceivers",
	posPointerDef:       "PointerDef",
	posNonStruct:        "NonStruct",
	posZeroSized:        "ZeroSized",
}

// String returns the string representation of a TypeProperty.
// If multiple flags are set, it returns a comma-separated list of names.
// If no flags are set, it returns "None".
func (e ErrorProperty) String() string {
	return bitflag.ToString(e, _errorPropertyNames[:], "None")
}
