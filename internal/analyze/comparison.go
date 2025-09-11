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
	"go/format"
	"go/token"
	"go/types"
	"strings"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// comparison analyzes a comparison operation (either binary like `==` or
// a function call like `errors.Is`) to determine if one of the operands is
// the address of a composite literal or a new() call.
//
// It reports a diagnostic if such a comparison is found, providing additional context
// if the comparison involves zero-sized types.
func (p pass) comparison(n ast.Node, left, right ast.Expr, checkIs bool) {
	var (
		typ        types.Type // The type of T in a &T{} or new(T) operand
		canCompare bool
		isLeft     bool     // operand detected is on the left side of the comparison
		other      ast.Expr // isLeft ? right : left
	)

	// Determine if one of the operands is a new literal (&T{} or new(T)), check the left first.
	if tl, cl, ok := p.isNewOrNonComparable(left); ok {
		typ, canCompare, other, isLeft = tl, cl, right, true
	} else if tr, cr, ok := p.isNewOrNonComparable(right); ok {
		typ, canCompare, other, isLeft = tr, cr, left, false
	} else {
		return // Not a comparison we are interested in.
	}

	ptr, isPtr := types.Unalias(typ).(*types.Pointer)

	isUndefined := false

	if isPtr {
		// Determine if the comparison is with a zero-sized type and the other operand is not nil.
		if typeutil.ZeroSized(ptr.Elem(), 0) {
			otherType, ok := p.TypesInfo.Types[other]
			// In this case, the result is undefined.
			isUndefined = !ok || !otherType.IsNil()
		}
	}

	tag := "equ"

	// If the type implements the error interface, it may be a valid comparison
	// in the context of errors.Is, which has special unwrapping rules.
	if typeutil.HasErrorMethod(typ) {
		if checkIs && shouldSuppressDiagnostic(typ, isLeft) {
			return
		}

		tag = "cmp"
	}

	// Report diagnostic
	var typeName string
	if isPtr { // e.g. &T{} or new(T), the type is a pointer.
		typeName = types.TypeString(ptr.Elem(), types.RelativeTo(p.Pkg))
	} else {
		typeName = types.TypeString(typ, types.RelativeTo(p.Pkg))
	}

	otherStr := "<unknown>"

	var sb strings.Builder
	if format.Node(&sb, p.Fset, other) == nil {
		otherStr = sb.String()
	}

	var format string

	switch {
	case !canCompare:
		format = "Result of comparison of %q with non-comparable variable of type %q is always false. (et:%s)"

	case isUndefined:
		format = "Result of comparison of %q with address of new zero-sized variable of type %q is false or undefined. (et:%s)"

	default:
		format = "Result of comparison of %q with address of new variable of type %q is always false. (et:%s+)"
	}

	p.ReportRangef(n, format, otherStr, typeName, tag)
}

// isNewOrNonComparable checks if an expression `x` is one of the following:
// 1. The address of a new composite literal: `&T{...}`
// 2. A call to the built-in `new()` function: `new(T)`
// 3. A non-comparable composite literal: `T{...}` where T is not comparable.
// It returns the type of the expression, a boolean indicating if the expression is comparable, and a boolean for success.
func (p pass) isNewOrNonComparable(x ast.Expr) (typ types.Type, canCompare, ok bool) {
	switch e := ast.Unparen(x).(type) {
	case *ast.UnaryExpr:
		if e.Op != token.AND {
			return nil, false, false // not &...
		}

		if _, ok := ast.Unparen(e.X).(*ast.CompositeLit); !ok {
			return nil, false, false // not &...{}
		}

		tv, ok := p.TypesInfo.Types[e]

		return tv.Type, true, ok

	case *ast.CallExpr:
		if len(e.Args) != 1 {
			return nil, false, false // some function
		}

		fun, ok := ast.Unparen(e.Fun).(*ast.Ident)
		if !ok || fun.Name != "new" {
			return nil, false, false // not new(...)
		}

		if _, ok := p.TypesInfo.Uses[fun].(*types.Builtin); !ok {
			return nil, false, false // not the built-in "new"
		}

		tv, ok := p.TypesInfo.Types[e]

		return tv.Type, true, ok

	case *ast.CompositeLit:
		if tv, ok := p.TypesInfo.Types[e.Type]; ok && !types.Comparable(tv.Type) {
			return tv.Type, false, true
		}

		return nil, false, false

	default:
		return nil, false, false
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
	if types.Implements(typ, errorIsInterface) {
		return true
	}

	// 2. `err.Unwrap()`: If `err` is the newly created literal (`isLeft` is true),
	//    and its type `*T` implements an `Unwrap` method, `errors.Is` will traverse
	//    the unwrapped errors. The comparison might then be valid against an unwrapped error.
	//    Thus, we suppress the diagnostic in this case.
	if isLeft &&
		(types.Implements(typ, errorUnwrapInterface) ||
			types.Implements(typ, errorUnwrapArrayInterface)) {
		return true
	}

	// We do not have dynamic runtime types, these heuristics rely on static type information
	// and seem to work well in practice.

	return false
}
