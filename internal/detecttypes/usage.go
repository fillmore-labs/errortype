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

package detecttypes

import (
	"context"
	"go/ast"
	"go/token"
	"go/types"
	"runtime/trace"

	"fillmore-labs.com/errortype/internal/detecttypes/properties"
	"fillmore-labs.com/errortype/internal/detectwrappers/result"
	"fillmore-labs.com/errortype/internal/errorconv"
	"fillmore-labs.com/errortype/internal/errorlike"
	"fillmore-labs.com/errortype/internal/printf"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// processUsage processes generic instantiations and all function declarations
// in the current package, visiting their bodies to perform error usage analysis.
func (d dpass) processUsage(ctx context.Context, errorFuncs result.ErrorFuncs) {
	defer trace.StartRegion(ctx, "usage").End()

	d.processInstances(errorFuncs)

	for f := range typeutil.AllFuncDecls(d.Files) {
		d.walkFunctionBody(f, errorFuncs)
	}
}

// processInstances inspects generic function and type instantiations where a
// type parameter constraint implements the error interface.
//
// Instantiations of detected [result.WrapperAsType] are recorded as targets
// rather than general instances to avoid double counting.
func (d dpass) processInstances(errorFuncs result.ErrorFuncs) {
	for arg := range errorconv.TypeArgs(d.TypesInfo) {
		if errorFuncs.IsAsTypeTarget(arg.Obj, arg.Index) {
			d.recordProperty(arg.Type, properties.PointerTarget, properties.ValueTarget)
			continue
		}

		d.recordProperty(arg.Type, properties.PointerInstance, properties.ValueInstance)
	}
}

// walkFunctionBody walks the function body with a usage visitor to analyze error usage.
func (d dpass) walkFunctionBody(f *ast.FuncDecl, errorFuncs result.ErrorFuncs) {
	var sig *types.Signature
	if fun, ok := d.TypesInfo.Defs[f.Name].(*types.Func); ok {
		sig = fun.Signature()
	}

	d.walkBlock(sig, f.Body, errorFuncs)
}

func (d dpass) walkBlock(sig *types.Signature, body *ast.BlockStmt, errorFuncs result.ErrorFuncs) {
	var results *types.Tuple
	if sig != nil {
		results = sig.Results()
	}

	if results != nil {
		for v := range results.Variables() {
			if typ := v.Type(); !types.IsInterface(typ) && errorlike.Implements(typ) {
				if tn, ptr, ok := typeutil.CanonicalTypeNameOf(typ); ok {
					prop := properties.ValueReturn
					if ptr {
						prop = properties.PointerReturn
					}

					d.addTypePropertyInCurrentPackage(tn, prop)
				}
			}
		}
	}

	if body == nil {
		return
	}

	v := usageVisitor{
		dpass:      d,
		errorFuncs: errorFuncs,
		results:    results,
	}

	ast.Walk(v, body)
}

type usageVisitor struct {
	dpass
	errorFuncs result.ErrorFuncs
	results    *types.Tuple
}

// walkExprs applies the visitor to each expression in the given list.
func (v usageVisitor) walkExprs(exprs ...ast.Expr) {
	for _, expr := range exprs {
		if expr != nil {
			ast.Walk(v, expr)
		}
	}
}

func (v usageVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.AssignStmt:
		v.handleAssign(n)

		return nil // Handled, stop descending.

	case *ast.ValueSpec:
		v.handleValueSpec(n)

		return nil // Handled, stop descending.

	case *ast.ReturnStmt:
		v.handleReturn(n)

		return nil // Handled, stop descending.

	case *ast.SendStmt:
		v.handleSend(n)

		return nil // Handled, stop descending.

	case *ast.FuncLit:
		// A function literal defines a new function.
		// We inspect its body for how it returns error types.
		sig, _ := v.TypesInfo.Types[n].Type.(*types.Signature)
		v.walkBlock(sig, n.Body, v.errorFuncs)

		return nil // Handled, stop descending.

	case *ast.CallExpr:
		if tv, ok := v.TypesInfo.Types[n.Fun]; ok && tv.IsType() {
			v.walkExprs(n.Args...)

			if errorlike.IsInterface(tv.Type) {
				// Explicit conversion to error interface, e.g., error(e).
				for argType := range errorconv.CallArgs(v.TypesInfo, n) {
					v.recordAssignProperty(argType)
				}
			} else {
				// Type cast, e.g., type MyError string; MyError("error").
				v.handleCast(tv.Type)
			}

			return nil
		}

		v.handleCallExpr(n)

		return nil // Handled, stop descending.

	case *ast.UnaryExpr:
		// Handle address-of composite literal, e.g., &MyError{}.
		if n.Op == token.AND {
			if cl, ok := ast.Unparen(n.X).(*ast.CompositeLit); ok {
				v.handleCompositeLit(cl, true)

				return nil // Handled, stop descending.
			}
		}

		return v // Continue visiting other unary expressions.

	case *ast.CompositeLit:
		// Handle value literals, e.g., MyError{}.
		v.handleCompositeLit(n, false)

		return nil // Handled, stop descending.

	case *ast.TypeAssertExpr:
		// Handle type assertions, e.g., err.(MyError).
		v.handleTypeAssert(n)

		if n.X != nil {
			ast.Walk(v, n.X)
		}

		return nil // Handled, stop descending.

	case *ast.TypeSwitchStmt:
		v.handleTypeSwitch(n)

		return nil // Handled, stop descending.

	case *ast.TypeSpec, *ast.CommentGroup:
		return nil // Skip local type declarations and comments.

	default:
		// Continue visiting other nodes.
		return v
	}
}

// handleValueSpec processes var and const statements.
func (v usageVisitor) handleValueSpec(n *ast.ValueSpec) {
	if len(n.Values) == 0 {
		return // uninitialized variables
	}

	if n.Type == nil {
		// Analyze initial values in `var` declarations.
		v.walkExprs(n.Values...)
		return
	}

	if tv, ok := v.TypesInfo.Types[n.Type]; !ok || !errorlike.IsInterface(tv.Type) {
		v.walkExprs(n.Values...)
		return
	}

	for _, val := range n.Values {
		if tv, ok := v.TypesInfo.Types[val]; ok {
			v.recordAssignProperty(tv.Type)
		}
	}
}

// handleAssign processes assignment statements (`=` and `:=`).
func (v usageVisitor) handleAssign(n *ast.AssignStmt) {
	v.walkExprs(n.Lhs...)
	v.walkExprs(n.Rhs...)

	var lhsType func(info *types.Info, expr ast.Expr) types.Type

	switch n.Tok {
	case token.ASSIGN:
		lhsType = typeOfExpr

	case token.DEFINE:
		// redeclarations can only appear in a multi-variable short declaration.
		if len(n.Lhs) < 2 {
			return
		}

		lhsType = typeOfRedeclared

	default:
		return
	}

	for i, lhs := range n.Lhs {
		if lTyp := lhsType(v.TypesInfo, lhs); lTyp == nil || !errorlike.IsInterface(lTyp) {
			continue
		}

		if tRHS, _ := typeutil.ResultOf(v.TypesInfo, n.Rhs, i); tRHS != nil {
			v.recordAssignProperty(tRHS)
		}
	}
}

func typeOfExpr(info *types.Info, expr ast.Expr) types.Type {
	if tv, ok := info.Types[expr]; ok {
		return tv.Type
	}

	return nil
}

func typeOfRedeclared(info *types.Info, expr ast.Expr) types.Type {
	id, ok := expr.(*ast.Ident)
	if !ok || id.Name == "_" {
		return nil
	}

	obj, ok := info.Uses[id]
	if !ok {
		return nil
	}

	return obj.Type()
}

// handleSend processes channel send statements (`ch <- val`).
func (v usageVisitor) handleSend(n *ast.SendStmt) {
	v.walkExprs(n.Chan, n.Value)

	if valType, _, ok := errorconv.SendValue(v.TypesInfo, n); ok {
		v.recordAssignProperty(valType)
	}
}

// handleTypeAssert processes a type assertion expression, e.g., v.(T).
func (d dpass) handleTypeAssert(n *ast.TypeAssertExpr) {
	if n.Type == nil {
		return // This is a type switch, not an assertion.
	}

	tv, ok := d.TypesInfo.Types[n.Type]
	if !ok {
		return // !ok means errors in type parsing
	}

	if !tv.IsType() {
		d.LogErrorf(n.Type, "expected type in assertion, got %#v", tv)

		return
	}

	// We can only analyze named types.
	tn, ptr, ok := typeutil.CanonicalTypeNameOf(tv.Type)
	if !ok {
		return
	}

	prop := properties.ValueAssert
	if ptr {
		prop = properties.PointerAssert
	}

	d.addTypePropertyInCurrentPackage(tn, prop)
}

// handleTypeSwitch processes a type switch statement, e.g., `switch v.(type) { case T: }`.
func (v usageVisitor) handleTypeSwitch(n *ast.TypeSwitchStmt) {
	if n.Init != nil {
		ast.Walk(v, n.Init)
	}

	if n.Assign != nil {
		ast.Walk(v, n.Assign)
	}

	if n.Body == nil {
		return
	}

	for _, stmt := range n.Body.List {
		clause, ok := stmt.(*ast.CaseClause)
		if !ok {
			continue
		}

		for _, caseExpr := range clause.List {
			tv, ok := v.TypesInfo.Types[caseExpr]
			if !ok || tv.IsNil() || !tv.IsType() {
				continue
			}

			tn, ptr, ok := typeutil.CanonicalTypeNameOf(tv.Type)
			if !ok {
				continue
			}

			prop := properties.ValueAssert
			if ptr {
				prop = properties.PointerAssert
			}

			v.addTypePropertyInCurrentPackage(tn, prop)
		}

		for _, bodyStmt := range clause.Body {
			ast.Walk(v, bodyStmt)
		}
	}
}

// handleCast processes a type conversion, e.g., T(v).
func (d dpass) handleCast(typ types.Type) {
	// We can only analyze named types.
	tn, ptr, ok := typeutil.CanonicalTypeNameOf(typ)
	if !ok {
		return
	}

	prop := properties.ValueCast
	if ptr {
		prop = properties.PointerCast
	}

	d.addTypePropertyInCurrentPackage(tn, prop)
}

// handleCompositeLit processes a composite literal, e.g., T{} or &T{}.
func (v usageVisitor) handleCompositeLit(n *ast.CompositeLit, isAddrOf bool) {
	v.walkExprs(n.Elts...)

	for valType := range errorconv.CompositeLitElts(v.TypesInfo, n) {
		v.recordAssignProperty(valType)
	}

	litTyp, ok := v.TypesInfo.Types[n]
	if !ok {
		return
	}

	litType := litTyp.Type

	var (
		tn             *types.TypeName
		ptr, namedType bool
	)

	if tn, ptr, namedType = typeutil.CanonicalTypeNameOf(litType); namedType {
		ptr = ptr || isAddrOf
	} else {
		if p, ok := types.Unalias(litType).(*types.Pointer); ok {
			litType = p.Elem()
		}

		// The core type also covers literals of type parameter type, e.g., S{} with S ~[]T.
		switch t := typeutil.CoreType(litType).(type) {
		case *types.Slice, *types.Array, *types.Map:
			elem := t.(interface{ Elem() types.Type }).Elem()
			tn, ptr, namedType = typeutil.CanonicalTypeNameOf(elem)
		}
	}

	if !namedType {
		return // Not a named type.
	}

	property := properties.ValueLiteral
	if ptr {
		property = properties.PointerLiteral
	}

	v.addTypePropertyInCurrentPackage(tn, property)
}

// handleReturn processes returned values, including multivalued returns.
func (v usageVisitor) handleReturn(ret *ast.ReturnStmt) {
	v.walkExprs(ret.Results...)

	if v.results == nil || len(ret.Results) == 0 {
		return
	}

	for i := range v.results.Len() {
		if !errorlike.Implements(v.results.At(i).Type()) {
			continue
		}

		resType, _ := typeutil.ResultOf(v.TypesInfo, ret.Results, i)
		if resType == nil || resType == types.Typ[types.UntypedNil] || types.IsInterface(resType) {
			continue
		}

		tn, ptr, ok := typeutil.CanonicalTypeNameOf(resType)
		if !ok {
			continue // Not a named type.
		}

		property := properties.ValueReturn
		if ptr {
			property = properties.PointerReturn
		}

		v.addTypePropertyInCurrentPackage(tn, property)
	}
}

// errorWrapper analyzes a function call to determine if it matches patterns like errors.As
// and identifies the target error type.
// It returns the target type or nil if the call is not of interest.
func (v usageVisitor) errorWrapper(call *ast.CallExpr) types.Type {
	fun, ef, ok := v.errorFuncs.Lookup(v.TypesInfo, call)
	if !ok {
		return nil // Not a known error wrapper.
	}

	if ef.Type == result.WrapperAsType {
		return nil // Handled in processInstances.
	}

	target, ok := ef.TargetType(v.TypesInfo, fun, call.Args)
	if !ok {
		return nil
	}

	if ef.Type == result.FuncAssert && !errorlike.Implements(target) {
		return nil
	}

	return target
}

func (v usageVisitor) handleCallExpr(call *ast.CallExpr) {
	if typeutil.BuiltinNew(v.TypesInfo, call) {
		v.handleBuiltinNew(call.Args[0])

		return
	}

	v.recordWrappedOperands(call)

	if target := v.errorWrapper(call); target != nil {
		tn, ptr, ok := typeutil.CanonicalTypeNameOf(target)
		if !ok {
			return // Not a named type.
		}

		property := properties.ValueTarget
		if ptr {
			property = properties.PointerTarget
		}

		v.addTypePropertyInCurrentPackage(tn, property)

		return
	}

	// not an error wrapper
	ast.Walk(v, call.Fun)
	v.walkExprs(call.Args...)

	for argType := range errorconv.CallArgs(v.TypesInfo, call) {
		v.recordAssignProperty(argType)
	}
}

func (v usageVisitor) handleBuiltinNew(arg ast.Expr) {
	switch e := ast.Unparen(arg).(type) {
	case *ast.CompositeLit:
		// treat new(MyError{...}) like &MyError{...}.
		v.handleCompositeLit(e, true)
		return

	case *ast.CallExpr:
		tv, ok := v.TypesInfo.Types[e.Fun]
		if !ok || !tv.IsType() {
			break
		}

		v.walkExprs(e.Args...)

		// treat new(myIntError(1)) like (*myIntError)(&1).
		if tn, ptr, ok := typeutil.CanonicalTypeNameOf(tv.Type); ok && !ptr {
			v.addTypePropertyInCurrentPackage(tn, properties.PointerCast)
		}

		return

	default:
	}

	ast.Walk(v, arg)
}

// recordWrappedOperands records the operands wrapped by %w in a call to a fmt.Errorf-like
// function, which end up in the error chain, as [properties.PointerWrap] or [properties.ValueWrap].
// The operands are walked separately, as for any call.
func (v usageVisitor) recordWrappedOperands(call *ast.CallExpr) {
	fun, ef, ok := v.errorFuncs.Lookup(v.TypesInfo, call)
	if !ok || ef.Type != result.WrapperErrorf || call.Ellipsis.IsValid() {
		return // Not fmt.Errorf-like, or the operands are passed as a slice.
	}

	formatExpr, ok := ef.SourceArg(call.Args, fun.MethodExpr)
	if !ok {
		return
	}

	for expr := range printf.WrappedOperands(v.TypesInfo, formatExpr, ef.TargetArgs(call.Args, fun.MethodExpr)) {
		if tv, ok := v.TypesInfo.Types[expr]; ok {
			v.recordProperty(tv.Type, properties.PointerWrap, properties.ValueWrap)
		}
	}
}

// recordAssignProperty analyzes the given type to determine if it's a pointer or
// value error and records a [properties.PointerAssign] or [properties.ValueAssign] property respectively.
func (d dpass) recordAssignProperty(typ types.Type) {
	d.recordProperty(typ, properties.PointerAssign, properties.ValueAssign)
}

// recordProperty records pointerProp or valueProp for the named error type of typ,
// depending on whether typ is a pointer. Interfaces and untyped nil are ignored.
func (d dpass) recordProperty(typ types.Type, pointerProp, valueProp properties.ErrorProperty) {
	if typ == nil || typ == types.Typ[types.UntypedNil] || types.IsInterface(typ) {
		return
	}

	tn, ptr, ok := typeutil.CanonicalTypeNameOf(typ)
	if !ok {
		return
	}

	prop := valueProp
	if ptr {
		prop = pointerProp
	}

	d.addTypePropertyInCurrentPackage(tn, prop)
}

// addTypePropertyInCurrentPackage sets a property on a type if it's a known error type
// in the current package and the property isn't yet set.
func (d dpass) addTypePropertyInCurrentPackage(tn *types.TypeName, property properties.ErrorProperty) {
	if !d.inCurrentPkg(tn) {
		return // Only relevant for types defined in the current package
	}

	if old, ok := d.ErrorTypes[tn]; ok && !old.Has(property) { // known, but property isn't set.
		d.ErrorTypes[tn] |= property
	}
}
