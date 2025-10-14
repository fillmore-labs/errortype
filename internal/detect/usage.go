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
	"context"
	"go/ast"
	"go/token"
	"go/types"
	"runtime/trace"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// processUsage processes all function declarations in the current package,
// visiting their bodies to perform error usage analysis.
func (p pass) processUsage(ctx context.Context) {
	defer trace.StartRegion(ctx, "usage").End()

	for f := range p.AllFuncDecls {
		p.walkFunctionBody(f)
	}
}

// walkFunctionBody walks the function body with a usage visitor to analyze error usage.
func (p pass) walkFunctionBody(f *ast.FuncDecl) {
	if f.Body == nil {
		return
	}

	v := usageVisitor{
		pass:       p,
		lastResult: typeutil.HasErrorResult(p.TypesInfo, f.Type.Results),
		inExpr:     false,
	}
	ast.Walk(v, f.Body)
}

type usageVisitor struct {
	pass
	lastResult int
	inExpr     bool // true when visiting expressions in assignment/return contexts
}

// walkExprs applies the visitor in expression context to each expression in the given list.
// This is used to analyze expressions in assignment, return, or declaration contexts.
func (v usageVisitor) walkExprs(exprs ...ast.Expr) {
	v.inExpr = true

	for _, expr := range exprs {
		ast.Walk(v, expr)
	}
}

func (v usageVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.AssignStmt:
		// Analyze the right-hand side of `:=` and `=` assignments.
		v.walkExprs(n.Rhs...)

		return nil // Expressions handled, stop descending.

	case *ast.ValueSpec:
		// Analyze initial values in `var` declarations.
		v.walkExprs(n.Values...)

		return nil // Expressions handled, stop descending.

	case *ast.ReturnStmt:
		v.handleReturn(n)

		return nil // Expressions handled, stop descending.

	case *ast.SendStmt:
		// Analyze the value sent to a channel.
		v.walkExprs(n.Value)

		return nil // Expressions handled, stop descending.

	case *ast.FuncLit:
		// A function literal defines a new function.
		// We inspect its body for how it returns error types.
		u := usageVisitor{
			pass:       v.pass,
			lastResult: typeutil.HasErrorResult(v.TypesInfo, n.Type.Results),
			inExpr:     false,
		}

		return u

	case *ast.CallExpr:
		if tv, ok := v.TypesInfo.Types[n.Fun]; ok && tv.IsType() {
			// Type cast, e.g., type MyError string; MyError("error").
			v.handleCast(tv.Type)

			return nil
		}

		v.handleCallExpr(n)

		return nil // Expressions handled, stop descending.

	case *ast.UnaryExpr:
		if !v.inExpr || n.Op != token.AND {
			return v // Continue for other unary expressions.
		}

		// Handle address-of operator in expression context, e.g., &MyError{}.
		cl, ok := ast.Unparen(n.X).(*ast.CompositeLit)
		if !ok {
			return v
		}

		v.handleCompositeLit(cl, true)

		return nil // Handled, stop descending.

	case *ast.CompositeLit:
		if !v.inExpr {
			return v // Continue if not in expression context.
		}

		// Handle value literals in expression context, e.g., MyError{}.
		v.handleCompositeLit(n, false)

		return nil // Handled, stop descending.

	case *ast.TypeAssertExpr:
		if !v.inExpr {
			return v // Continue if not in expression context.
		}

		// Handle type assertions in expression context, e.g., err.(MyError).
		v.handleTypeAssert(n)

		return nil // Handled, stop descending.

	case ast.Expr:
		// We have handled all expression contexts we are interested in.
		// Skip any other expressions to avoid redundant analysis.
		return nil

	case *ast.CommentGroup:
		return nil // Skip comments

	default:
		// Continue visiting other nodes (e.g., statements).
		return v
	}
}

// handleTypeAssert processes a type assertion expression, e.g., v.(T).
func (p pass) handleTypeAssert(n *ast.TypeAssertExpr) {
	if n.Type == nil {
		return // This is a type switch, not an assertion.
	}

	tv, ok := p.TypesInfo.Types[n.Type]
	if !ok {
		return // !ok means errors in type parsing
	}

	if !tv.IsType() {
		p.LogErrorf(n.Type, "Expected type in assertion, got %#v", tv)

		return
	}

	// We can only analyze named types.
	tn, ptr, ok := typeutil.TypeNameOf(tv.Type)
	if !ok {
		return
	}

	prop := ValueAssert
	if ptr {
		prop = PointerAssert
	}

	p.addTypePropertyInCurrentPackage(tn, prop)
}

// handleCast processes a type conversion, e.g., T(v).
func (p pass) handleCast(typ types.Type) {
	// We can only analyze named types.
	tn, ptr, ok := typeutil.TypeNameOf(typ)
	if !ok {
		return
	}

	prop := ValueCast
	if ptr {
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

	var (
		tn             *types.TypeName
		ptr, namedType bool
	)

	switch t := tv.Type.(type) {
	case *types.Slice:
		tn, ptr, namedType = typeutil.TypeNameOf(t.Elem())

	case *types.Array:
		tn, ptr, namedType = typeutil.TypeNameOf(t.Elem())

	case *types.Map:
		tn, ptr, namedType = typeutil.TypeNameOf(t.Elem())

	default:
		tn, _, namedType = typeutil.TypeNameOf(t)
		ptr = isAddrOf
	}

	if !namedType {
		return // Not a named type.
	}

	property := ValueLiteral
	if ptr {
		property = PointerLiteral
	}

	p.addTypePropertyInCurrentPackage(tn, property)
}

// handleReturn processes returned values, T{} or &T{}.
func (v usageVisitor) handleReturn(ret *ast.ReturnStmt) {
	v.walkExprs(ret.Results...)

	if v.lastResult < 0 || len(ret.Results) <= v.lastResult {
		return
	}

	res := ret.Results[v.lastResult]

	resType, ok := v.TypesInfo.Types[res]
	if !ok || resType.IsNil() {
		return // !ok means errors in type parsing, nil is fine.
	}

	if !resType.IsValue() { // should not happen
		v.LogErrorf(ret, "Expected returned value in %d , got %#v", v.lastResult, resType)
	}

	tn, ptr, ok := typeutil.TypeNameOf(resType.Type)
	if !ok {
		return // Not a named type.
	}

	property := ValueReturn
	if ptr {
		property = PointerReturn
	}

	v.addTypePropertyInCurrentPackage(tn, property)
}

// handleErrorAs analyzes a function call to determine if it matches patterns like errors.As and identifies the target argument.
// It returns the target argument, or nil if the function is not of interest.
func (p pass) handleErrorAs(n *ast.CallExpr) (target types.Type, ok bool) {
	fun, typeParams, methodExpr, ok := typeutil.FuncOf(p.TypesInfo, n.Fun)
	if !ok {
		return nil, false // Could not resolve function, might be a func variable.
	}

	// errorsAs maps a function name to the index of its "target" argument.
	funcName := typeutil.FuncNameOf(fun)

	asfunc, ok := typeutil.KnownFuncs[funcName]
	if !ok || asfunc.Kind() != typeutil.KindAs {
		return nil, false // Not a function we are interested in.
	}

	targetArgIndex, typeParam := asfunc.AsTarget()
	if typeParam >= 0 { // Handle generic functions like `errors.AsType[T]`.
		if len(typeParams) > typeParam {
			typ := typeParams[typeParam]
			if tv, ok := p.TypesInfo.Types[typ]; ok && tv.IsType() && typeutil.HasErrorMethod(tv.Type) {
				return tv.Type, ok
			}
		}
	}

	if targetArgIndex < 0 {
		return nil, false
	}

	if methodExpr {
		// For method expression calls ("(*assert.Assertions).ErrorsAs(a, ...)"),
		// the receiver `a` is the first argument. The argument indices in `errorsAs`
		// are for the function form, so we increment the index to correctly locate
		// the target argument in the method call expression.
		targetArgIndex++
	}

	if len(n.Args) <= targetArgIndex {
		return nil, true // Maybe called with the result of a multivalued function
	}

	targetArg := n.Args[targetArgIndex]

	typ, ok := p.TypesInfo.Types[targetArg]
	if !ok {
		return nil, true // !ok means errors in type parsing
	}

	ptr, ok := typ.Type.Underlying().(*types.Pointer)
	if !ok {
		return nil, true
	}

	return ptr.Elem(), true
}

func (v usageVisitor) handleCallExpr(n *ast.CallExpr) {
	if target, ok := v.handleErrorAs(n); ok {
		tn, ptr, ok := typeutil.TypeNameOf(target)
		if !ok {
			return // Not a named type.
		}

		property := ValueTarget
		if ptr {
			property = PointerTarget
		}

		v.addTypePropertyInCurrentPackage(tn, property)

		return
	}

	// not an `errors.As`-like function
	v.walkExprs(n.Args...)

	// TODO: handle target arguments of `errors.Is`

	if f, ok := n.Fun.(*ast.FuncLit); ok { // For immediately invoked function literals, examine their body.
		u := usageVisitor{
			pass:       v.pass,
			lastResult: typeutil.HasErrorResult(v.TypesInfo, f.Type.Results),
			inExpr:     false,
		}

		ast.Walk(u, f.Body)
	}
}

// addTypePropertyInCurrentPackage sets a property on a type if it's a known error type
// in the current package and the property isn't yet set.
func (p pass) addTypePropertyInCurrentPackage(tn *types.TypeName, property ErrorProperty) {
	if !p.inCurrentPkg(tn) {
		return // Only relevant for types defined in the current package
	}

	old, ok := p.GetTypeProperty(tn)

	if ok && old&property != property { // known, but property isn't set.
		p.SetTypeProperty(tn, old|property)
	}
}
