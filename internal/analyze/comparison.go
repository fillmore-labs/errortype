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
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// comparison analyzes a comparison operation (either binary like `==` or
// a function call like `errors.Is`) to detect problematic comparisons.
//
// It checks for two types of issues:
// 1. Comparisons with new pointers (&T{} or new(T)) - always false
// 2. Comparisons with non-comparable types - always false
//
// Reports diagnostics for such comparisons, with special handling for
// zero-sized types and error interfaces.
func (p Pass) comparison(n analysis.Range, left, right ast.Expr, direct bool) {
	// First, check if one of the operands is a new pointer (&T{} or new(T)), checking the left first.
	switch {
	case p.isNewPointer(left):
		if tv, ok := p.TypesInfo.Types[left]; ok {
			typ, other, isLeft := tv.Type, right, true
			p.diagNewComparison(n, typ, other, direct, isLeft)
		}

		return

	case p.isNewPointer(right):
		if tv, ok := p.TypesInfo.Types[right]; ok {
			typ, other, isLeft := tv.Type, left, false
			p.diagNewComparison(n, typ, other, direct, isLeft)
		}

		return
	}

	// Check for uncomparable types
	if tv, ok := p.TypesInfo.Types[left]; ok && !typeutil.Comparable(tv.Type) {
		typ, other, isLeft := tv.Type, right, true
		p.diagUncomparable(n, typ, other, direct, isLeft)

		return
	}

	if tv, ok := p.TypesInfo.Types[right]; ok && !typeutil.Comparable(tv.Type) {
		typ, other, isLeft := tv.Type, left, false
		p.diagUncomparable(n, typ, other, direct, isLeft)

		return
	}

	// Not a comparison we are interested in.
}

// diagNewComparison reports diagnostics for comparisons with new pointers (&T{} or new(T)).
// These comparisons are always false since each new allocation creates a unique address.
func (p Pass) diagNewComparison(n analysis.Range, typ types.Type, other ast.Expr, direct, isLeft bool) {
	tn, ptr := types.Unalias(typ).(*types.Pointer)
	if !ptr { // should not happen
		p.ReportErrorf(n, "Expected pointer type, got %T", typ)

		return
	}

	isUndefined := false

	// Determine if the comparison is with a zero-sized type and the other operand is not nil.
	if typeutil.ZeroSized(tn.Elem()) {
		if otherType, ok := p.TypesInfo.Types[other]; ok && !otherType.IsNil() {
			// In this case, the result is undefined.
			isUndefined = true
		}
	}

	tag := "equ"
	if !direct {
		tag = "cmp"
	}

	// If the type implements the error interface, it may be a valid comparison
	// in the context of errors.Is, which has special unwrapping rules.
	if !direct && p.CheckIs() && typeutil.HasErrorMethod(typ) && shouldSuppressDiagnostic(typ, isLeft) {
		return
	}

	// Report diagnostic
	var msg string

	switch {
	case isUndefined:
		msg = "Result of comparison of %q with address of new zero-sized variable of type %q is false or undefined. (et:%s)"

	default:
		msg = "Result of comparison of %q with address of new variable of type %q is always false. (et:%s+)"
	}

	otherName := p.exprToString(other)
	typeName := types.TypeString(tn.Elem(), types.RelativeTo(p.Pkg))

	p.ReportRangef(n, msg, otherName, typeName, tag)
}

// diagUncomparable reports diagnostics for comparisons with non-comparable types.
// These comparisons are always false or panic.
func (p Pass) diagUncomparable(n analysis.Range, typ types.Type, other ast.Expr, direct, isLeft bool) {
	if otherType, ok := p.TypesInfo.Types[other]; ok && otherType.IsNil() {
		return
	}

	tag := "equ"
	if !direct {
		tag = "cmp"
	}

	// If the type implements the error interface, it may be a valid comparison
	// in the context of errors.Is, which has special unwrapping rules.
	if !direct && p.CheckIs() && typeutil.HasErrorMethod(typ) && shouldSuppressDiagnostic(typ, isLeft) {
		return
	}

	// Report diagnostic
	const msg = "Result of comparison of %q with non-comparable variable of type %q is always false. (et:%s+)"
	otherName := p.exprToString(other)
	typeName := types.TypeString(typ, types.RelativeTo(p.Pkg))

	p.ReportRangef(n, msg, otherName, typeName, tag)
}

// isNewPointer checks if an expression `x` is one of the following:
// 1. The address of a new composite literal: `&T{...}`
// 2. A call to the built-in `new()` function: `new(T)`
// It returns the type of the expression and a boolean for success.
func (p Pass) isNewPointer(x ast.Expr) bool {
	switch e := ast.Unparen(x).(type) {
	case *ast.UnaryExpr:
		if e.Op != token.AND {
			return false // not &...
		}

		if _, ok := ast.Unparen(e.X).(*ast.CompositeLit); !ok {
			return false // not &...{}
		}

		return true

	case *ast.CallExpr:
		return typeutil.BuiltinNew(p.TypesInfo, e)

	default:
		return false
	}
}

// shouldSuppressDiagnostic determines whether a diagnostic should be suppressed.
// This is primarily relevant for `errors.Is` calls, where certain patterns involving
// `Is` or `Unwrap` methods might make the comparison legitimate despite involving a new address.
func shouldSuppressDiagnostic(typ types.Type, isLeft bool) bool {
	// The standard library `errors.Is(err, target)` function checks if `err` (or an error
	// in its `Unwrap` tree) matches `target`. This matching can occur in several ways.
	// For this linter, which flags `errors.Is(err, &T{})` (where `&T{}` is the `target`),
	// we are concerned with two scenarios for suppression:

	// 1. Methods are called on the error tree of the first argument only, not the target.
	//    Since we do not have dynamic runtime types, we rely on heuristics and assume when
	//    `target` could be matched by an `Is(error) bool` method of `err`, it would be the
	//    `Is(error) bool` method of `target` and suppress the diagnostic in this case.
	if typeutil.HasMethod(typ, "Is", typeutil.HasIsSig) {
		return true
	}

	// 2. `err.Unwrap()`: If `err` is the newly created literal (`isLeft` is true),
	//    and its type `*T` implements an `Unwrap` method, `errors.Is` will traverse
	//    the unwrapped errors. The comparison might then be valid against an unwrapped error.
	//    Thus, we suppress the diagnostic in this case.
	if isLeft && typeutil.HasMethod(typ, "Unwrap", typeutil.HasUnwrapSig) {
		return true
	}

	// We do not have dynamic runtime types, these heuristics rely on static type information
	// and seem to work well in practice.

	return false
}
