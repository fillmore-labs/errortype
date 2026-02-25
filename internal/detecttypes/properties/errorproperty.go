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

import "fillmore-labs.com/errortype/internal/detecttypes/result"

// ErrorProperty is a bitmask representing properties of an error type's definition
// and usage. These properties are collected to determine whether a type should be
// consistently used as a pointer (*T) or a value (T).
//
//go:generate go tool bitmask -type ErrorProperty
type ErrorProperty uint64

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
	// EmbeddedPointerMethod is set if such a method is not directly defined on the type, but embedded.
	EmbeddedPointerMethod // EmbeddedPointerMethod

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

	// --- Properties from compile-time assertions (e.g., `var _ error = ...`) ---.

	// PointerDecl is set for a pointer assertion, e.g., `var _ error = (*T)(nil)`.
	PointerDecl // PointerDecl
	// ValueDecl is set for a value assertion, e.g., `var _ error = T{}`.
	ValueDecl // ValueDecl

	// --- Properties from variable declarations (e.g., `var ErrSomething = ...`) ---.

	// PointerVar is set for pointer usage, e.g., `var Err = &T{}`.
	PointerVar // PointerVar
	// ValueVar is set for value usage, e.g., `var Err = T{}`.
	ValueVar // ValueVar

	// --- Properties from usage in return statements ---.

	// PointerReturn is set for pointer usage, e.g., `return &T{}`.
	PointerReturn // PointerReturn
	// ValueReturn is set for value usage, e.g., `return T{}`.
	ValueReturn // ValueReturn

	// --- Properties from operands wrapped by %w in fmt.Errorf-like functions ---.

	// PointerWrap is set for pointer usage, e.g., `fmt.Errorf("...: %w", &T{})`.
	PointerWrap // PointerWrap
	// ValueWrap is set for value usage, e.g., `fmt.Errorf("...: %w", T{})`.
	ValueWrap // ValueWrap

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

	// --- Properties from conversions to error interfaces ---.

	// PointerAssign is set when a pointer *T is converted to an error interface (e.g., in assignments, call arguments, channel sends, or composite literals).
	PointerAssign // PointerAssign
	// ValueAssign is set when a value T is converted to an error interface (e.g., in assignments, call arguments, channel sends, or composite literals).
	ValueAssign // ValueAssign

	// --- Properties from generic type instantiations ---.

	// PointerInstance is set when a pointer *T instantiates a type parameter constrained to an error-like interface.
	PointerInstance // PointerInstance
	// ValueInstance is set when a value T instantiates a type parameter constrained to an error-like interface.
	ValueInstance // ValueInstance

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

	// --- Count ---.

	NumProperties = iota

	// --- Others ---.

	// OverrideMask identifies any override property.
	OverrideMask = PointerOverride | ValueOverride | SuppressOverride

	// DeclMask identifies compile-time assertions.
	DeclMask = PointerDecl | ValueDecl
)

// _propertyPairs defines the categories of evidence used to determine if an error
// type is a pointer or value type. The pairs are ordered by precedence, from
// the strongest evidence to the weakest. The first category with a non-contradictory
// signal determines the type.
//
// Compile-time assertions like `var _ error = (*T)(nil)` exist only to declare intent, so they
// rank above sentinel errors, which are uses of the type that might be the mistake.
//
// Operands wrapped by %w rank with returns: `return fmt.Errorf("...: %w", T{})` is `return T{}`
// with added context, and both put T into the error chain that callers query.
//
// Conversions to error interfaces determine which dynamic type errors.As and type assertions
// will encounter, so they rank above generic type instantiations, composite literals, and casts,
// which only show how a value was instantiated or constructed. They rank below type assertions
// and errors.As-like targets, which state the expected type deliberately, while assignments
// are the sites that are checked. Generic instantiations rank above composite literals and casts
// because they commit a concrete form to an error-constrained type parameter.
var _propertyPairs = [...]struct{ pointerProp, valueProp ErrorProperty }{
	{PointerAlias, ValueAlias},         // Aliases of imported error types.
	{PointerDecl, ValueDecl},           // `var _ error` assertions, which only declare intent.
	{PointerVar, ValueVar},             // Sentinel errors.
	{PointerReturn, ValueReturn},       // Usage in `return` statements.
	{PointerWrap, ValueWrap},           // Operands wrapped by %w in fmt.Errorf-like functions.
	{PointerAssert, ValueAssert},       // Usage in type assertions and type switches.
	{PointerTarget, ValueTarget},       // Usage in errors.As-like functions.
	{PointerAssign, ValueAssign},       // Conversions to error interfaces.
	{PointerInstance, ValueInstance},   // Instantiations of error-constrained type parameters.
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
	switch e.Mask(OverrideMask) { // Overrides have the highest precedence.
	case SuppressOverride:
		return result.Suppress

	case PointerOverride:
		return result.Pointer

	case ValueOverride:
		return result.Value
	}

	// Errors with pointer receivers can only be used in one way.
	if e.Has(PointerReceiver) {
		return result.Pointer
	}

	// Errors with an `Unwrap() error` method with pointer receiver would behave differently as values.
	// An explicit value assertion contradicts that, so then the receiver is the mistake. It can be fixed
	// where the method is declared, so embedded methods still decide.
	if e.Has(PointerMethod) && (e.Mask(DeclMask) != ValueDecl || e.Has(EmbeddedPointerMethod)) {
		return result.Pointer
	}

	for _, pair := range _propertyPairs {
		switch pointerProp, valueProp := pair.pointerProp, pair.valueProp; e.Mask(pointerProp | valueProp) { // Check for a non-contradictory usage within this category.
		case pointerProp:
			return result.Pointer

		case valueProp:
			return result.Value
		}
	}

	// A special case for defined pointer types like `type T *S`.
	// Although the underlying type is a pointer, `T` itself is used as a value
	// (e.g., you return `T`, not `*T`), so we treat it as a value type.
	if e.Has(PointerDef) {
		return result.Value
	}

	// No unambiguous usage was found.
	return result.Undecided
}
