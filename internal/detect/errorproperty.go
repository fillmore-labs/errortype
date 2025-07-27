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

package detect

import (
	"fmt"
	"strings"

	"fillmore-labs.com/errortype/internal/errortypes"
)

// ErrorProperty is a bitmask representing properties of an error type's definition
// and usage. These properties are collected to determine whether a type should be
// consistently used as a pointer (*T) or a value (T).
type ErrorProperty int32

// Properties are grouped by the heuristic that discovers them.
const (
	// --- Properties from Error() method receiver ---.

	// PointerReceiver is set if the Error() method has a pointer receiver,
	// e.g., `func (e *MyError) Error() string`. This is a strong indicator.
	PointerReceiver ErrorProperty = 1 << iota

	// --- Properties from user-defined overrides ---.

	// SuppressOverride is set if usage checks for this type are explicitly suppressed.
	SuppressOverride
	// PointerOverride is set if the type is explicitly marked as a pointer type in overrides.
	PointerOverride
	// ValueOverride is set if the type is explicitly marked as a value type in overrides.
	ValueOverride

	// --- Properties from variable declarations (e.g., `var ErrSomething = ...`) ---.

	// PointerVar is set for pointer usage, e.g., `var _ error = &T{}` or `var Err = &T{}`.
	PointerVar
	// ValueVar is set for value usage, e.g., `var _ error = T{}` or `var Err = T{}`.
	ValueVar

	// --- Properties from type aliases (e.g., `type T = V`) ---.

	// PointerAlias is set for an alias to an imported pointer-type error.
	PointerAlias
	// ValueAlias is set for an alias to an imported value-type error.
	ValueAlias

	// --- Properties from usage in return statements ---.

	// PointerReturn is set for pointer usage, e.g., `return &T{}`.
	PointerReturn
	// ValueReturn is set for value usage, e.g., `return T{}`.
	ValueReturn

	// --- Properties from usage in type assertions ---.

	// PointerAssert is set for pointer usage, e.g., `err.(*T)`.
	PointerAssert
	// ValueAssert is set for value usage, e.g., `err.(T)`.
	ValueAssert

	// --- Properties from targets in errors.As-like functions ---.

	// PointerTarget is set for pointer usage, e.g., `val target *T; ... errors.As(err, &target)`.
	PointerTarget
	// ValueTarget is set for value usage, e.g., `val target T; ... errors.As(err, &target)`.
	ValueTarget

	// --- Properties from usage in composite literals ---.

	// PointerLiteral is set for pointer usage, e.g., `&T{}`.
	PointerLiteral
	// ValueLiteral is set for value usage, e.g., `T{}`.
	ValueLiteral

	// --- Properties from usage in type casts ---.

	// PointerCast is set for pointer usage, e.g., `(*T)(v)`.
	PointerCast
	// ValueCast is set for value usage, e.g., `(T)(v)`.
	ValueCast

	// --- Properties from other method receivers ---.

	// PointerReceivers is set if all methods on the type have pointer receivers (a weak indicator).
	PointerReceivers
	// ValueReceivers is set if all methods on the type have value receivers (a weak indicator).
	ValueReceivers

	// --- Properties from type definition ---.

	// PointerDef is set for defined pointer types, e.g., `type T *S`. Such types are used like values.
	PointerDef

	// NonStruct is set if the defined type is not a structure or pointer.
	NonStruct

	// --- Others ---.

	// None. We have no idea.
	None ErrorProperty = 0

	// OverrideMask is a bitmask to identify any override property.
	OverrideMask = PointerOverride | ValueOverride | SuppressOverride
)

var errorProperties = map[ErrorProperty]string{
	PointerReceiver:  "PointerReceiver",
	SuppressOverride: "SuppressOverride",
	PointerOverride:  "PointerOverride",
	ValueOverride:    "ValueOverride",
	PointerVar:       "PointerVar",
	ValueVar:         "ValueVar",
	PointerAlias:     "PointerAlias",
	ValueAlias:       "ValueAlias",
	PointerReturn:    "PointerReturn",
	ValueReturn:      "ValueReturn",
	PointerAssert:    "PointerAssert",
	ValueAssert:      "ValueAssert",
	PointerTarget:    "PointerTarget",
	ValueTarget:      "ValueTarget",
	PointerLiteral:   "PointerLiteral",
	ValueLiteral:     "ValueLiteral",
	PointerCast:      "PointerCast",
	ValueCast:        "ValueCast",
	PointerReceivers: "PointerReceivers",
	ValueReceivers:   "ValueReceivers",
	PointerDef:       "PointerDef",
	NonStruct:        "NonStruct",
}

// String returns the string representation of a TypeProperty.
// If multiple flags are set, it returns a comma-separated list of names.
// If no flags are set, it returns "None".
func (e ErrorProperty) String() string {
	if e == None {
		return "None"
	}

	var parts []string

	for flag := ErrorProperty(1); flag != 0; flag <<= 1 {
		if e&flag != 0 {
			name, ok := errorProperties[flag]
			if !ok {
				name = fmt.Sprintf("Unknown(%d)", int(flag))
			}

			parts = append(parts, name)
		}
	}

	return strings.Join(parts, ", ")
}

// propertyPairs defines the categories of evidence used to determine if an error
// type is a pointer or value type. The pairs are ordered by precedence, from
// the strongest evidence to the weakest. The first category with a non-contradictory
// signal determines the type.
var propertyPairs = [...][2]ErrorProperty{
	{PointerOverride, ValueOverride},   // Strongest: Explicit user override.
	{PointerVar, ValueVar},             // Sentinel errors or `var _ error` assertions.
	{PointerAlias, ValueAlias},         // Aliases of imported error types.
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
	if e&SuppressOverride != 0 { // Suppression override has the highest precedence.
		return errortypes.SuppressType
	}

	if e&PointerReceiver != 0 { // Errors with pointer receivers can only be used in only one way.
		return errortypes.PointerType
	}

	for _, pair := range propertyPairs {
		pointerProp, valueProp := pair[0], pair[1]
		switch e & (pointerProp | valueProp) { // Check for a non-contradictory usage within this category.
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
	return errortypes.Undecided
}
