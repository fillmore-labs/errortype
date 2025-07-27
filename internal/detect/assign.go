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
	"go/ast"
	"go/token"
	"go/types"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// assignVisitor wraps a pass to visit nodes in assignment or return contexts,
// tracking how error types are used.
//
// The split in two visitors has historic reasons, they probably should be merged.
type assignVisitor struct{ pass }

func (v assignVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.UnaryExpr:
		if n.Op != token.AND {
			return v // Continue for other unary expressions.
		}
		// Handle address-of operator, e.g., &MyError{}.
		cl, ok := ast.Unparen(n.X).(*ast.CompositeLit)
		if !ok {
			return v // Continue for other unary expressions.
		}

		v.handleCompositeLit(cl, true)

		return nil // Handled, stop descending.

	case *ast.CompositeLit:
		// Handle value literals, e.g., MyError{}.
		v.handleCompositeLit(n, false)

		return nil // Handled, stop descending.

	case *ast.TypeAssertExpr:
		// Handle type assertions, e.g., err.(MyError).
		v.handleTypeAssert(n)

		return nil // Handled, stop descending.

	case *ast.CallExpr:
		if tv, ok := v.TypesInfo.Types[n.Fun]; ok && tv.IsType() {
			// Type cast, e.g., type MyError string; MyError("error").
			v.handleCast(tv.Type)

			return nil
		}

		v.handleCallExpr(n)

		return nil // Handled, stop descending.

	case *ast.FuncLit:
		// A function literal in an assignment context defines a new function.
		// We need to inspect its body for how it returns error types.
		u := usageVisitor{
			pass:       v.pass,
			lastResult: typeutil.HasErrorResult(v.TypesInfo, n.Type.Results),
		}

		ast.Walk(u, n.Body)

		return nil // Handled, stop descending.

	default:
		// For all other nodes, continue visiting children.
		return v
	}
}

// handleTypeAssert processes a type assertion expression, e.g., v.(T).
func (p pass) handleTypeAssert(n *ast.TypeAssertExpr) {
	if n.Type == nil {
		return // This is a type switch, not an assertion.
	}

	tv := p.TypesInfo.Types[n.Type]
	if !tv.IsType() {
		p.LogErrorf(n.Type, "Expected type in assertion, got %#v", tv)

		return
	}

	// We can only analyze named types.
	tn, isPtr, ok := typeutil.TypeNameOf(tv.Type)
	if !ok {
		return
	}

	prop := ValueAssert
	if isPtr {
		prop = PointerAssert
	}

	p.addTypePropertyInCurrentPackage(tn, prop)
}

// handleCast processes a type conversion, e.g., T(v).
func (p pass) handleCast(typ types.Type) {
	// We can only analyze named types.
	tn, isPtr, ok := typeutil.TypeNameOf(typ)
	if !ok {
		return
	}

	prop := ValueCast
	if isPtr {
		prop = PointerCast
	}

	p.addTypePropertyInCurrentPackage(tn, prop)
}

// handleCompositeLit processes a composite literal, e.g., T{} or &T{}.
func (p pass) handleCompositeLit(n *ast.CompositeLit, isAddrOf bool) {
	if n.Type == nil {
		return // Within a composite literal of array, slice, or map
	}

	tv := p.TypesInfo.Types[n.Type]
	if !tv.IsType() {
		p.LogErrorf(n.Type, "Expected type in composite literal, got %#v", tv)

		return
	}

	tn, isPtr, ok := typeutil.TypeNameOf(tv.Type)
	if !ok {
		return // Not a named type.
	}

	if isPtr { // should not happen
		p.LogErrorf(n, "Composite literal of a pointer type '%s'", types.TypeString(tn.Type(), nil))

		return
	}

	property := ValueLiteral
	if isAddrOf {
		property = PointerLiteral
	}

	p.addTypePropertyInCurrentPackage(tn, property)
}
