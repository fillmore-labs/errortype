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
	"fillmore-labs.com/errortype/internal/typeutil"
)

// processUsage processes generic instantiations and all function declarations
// in the current package, visiting their bodies to perform error usage analysis.
func (d dpass) processUsage(ctx context.Context) {
	defer trace.StartRegion(ctx, "usage").End()

	d.processInstances()

	for f := range typeutil.AllFuncDecls(d.Files) {
		d.walkFunctionBody(f)
	}
}

// processInstances inspects generic function and type instantiations where a
// type parameter constraint implements the error interface.
func (d dpass) processInstances() {
	for id, inst := range d.TypesInfo.Instances {
		obj, ok := d.TypesInfo.Uses[id]
		if !ok {
			continue
		}

		t, ok := obj.Type().(interface{ TypeParams() *types.TypeParamList })
		if !ok {
			continue
		}

		tparams := t.TypeParams()
		for i := range min(tparams.Len(), inst.TypeArgs.Len()) {
			if !typeutil.IsInterfaceWithError(tparams.At(i).Constraint()) {
				continue
			}

			tn, ptr, ok := typeutil.TypeNameOf(inst.TypeArgs.At(i))
			if !ok {
				continue
			}

			prop := properties.ValueTarget
			if ptr {
				prop = properties.PointerTarget
			}

			d.addTypePropertyInCurrentPackage(tn, prop)
		}
	}
}

// walkFunctionBody walks the function body with a usage visitor to analyze error usage.
func (d dpass) walkFunctionBody(f *ast.FuncDecl) {
	var sig *types.Signature
	if fun, ok := d.TypesInfo.Defs[f.Name].(*types.Func); ok {
		sig = fun.Signature()
	}

	d.walkBlock(sig, f.Body)
}

func (d dpass) walkBlock(sig *types.Signature, body *ast.BlockStmt) {
	var results *types.Tuple
	if sig != nil {
		results = sig.Results()
	}

	if results != nil {
		for v := range results.Variables() {
			if typ := v.Type(); !types.IsInterface(typ) && typeutil.HasErrorMethod(typ) {
				if tn, ptr, ok := typeutil.TypeNameOf(typ); ok {
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
		dpass:   d,
		results: results,
	}

	ast.Walk(v, body)
}

type usageVisitor struct {
	dpass
	results *types.Tuple
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
		// Analyze initial values in `var` declarations.
		v.walkExprs(n.Values...)

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
		v.walkBlock(sig, n.Body)

		return nil // Handled, stop descending.

	case *ast.CallExpr:
		if tv, ok := v.TypesInfo.Types[n.Fun]; ok && tv.IsType() {
			v.walkExprs(n.Args...)

			if typeutil.IsInterfaceWithError(tv.Type) {
				// Explicit conversion to error interface, e.g., error(e).
				if len(n.Args) == 1 {
					if argType, ok := v.TypesInfo.Types[n.Args[0]]; ok {
						v.recordAssignProperty(argType.Type)
					}
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
		if n.Op != token.AND {
			return v // Continue for other unary expressions.
		}

		// Handle address-of operator, e.g., &MyError{}.
		if cl, ok := ast.Unparen(n.X).(*ast.CompositeLit); ok {
			v.handleCompositeLit(cl, true)

			return nil // Handled, stop descending.
		}

		v.handleAddrOf(n)

		return v

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

// handleAssign processes assignment statements (`=` and `:=`).
func (v usageVisitor) handleAssign(n *ast.AssignStmt) {
	v.walkExprs(n.Lhs...)
	v.walkExprs(n.Rhs...)

	switch n.Tok {
	case token.ASSIGN:
		for i, lhs := range n.Lhs {
			if tv, ok := v.TypesInfo.Types[lhs]; !ok || !typeutil.IsInterfaceWithError(tv.Type) {
				continue
			}

			tRHS, _ := typeutil.ResultOf(v.TypesInfo, n.Rhs, i)
			if tRHS == nil {
				continue
			}

			v.recordAssignProperty(tRHS)
		}

	case token.DEFINE:
		// redeclaration can only appear in a multi-variable short declaration.
		if len(n.Lhs) < 2 {
			return
		}

		for i, lhs := range n.Lhs {
			id, ok := lhs.(*ast.Ident)
			if !ok {
				continue
			}

			// `:= ` can change the type only on redeclarations.
			if obj, ok := v.TypesInfo.Uses[id]; !ok || !typeutil.IsInterfaceWithError(obj.Type()) {
				continue
			}

			tRHS, _ := typeutil.ResultOf(v.TypesInfo, n.Rhs, i)
			if tRHS == nil {
				continue
			}

			v.recordAssignProperty(tRHS)
		}
	}
}

// handleSend processes channel send statements (`ch <- val`).
func (v usageVisitor) handleSend(n *ast.SendStmt) {
	v.walkExprs(n.Chan, n.Value)

	chanType, ok := v.TypesInfo.Types[n.Chan]
	if !ok {
		return
	}

	if ch, ok := chanType.Type.Underlying().(*types.Chan); !ok || !typeutil.IsInterfaceWithError(ch.Elem()) {
		return
	}

	valType, ok := v.TypesInfo.Types[n.Value]
	if !ok {
		return
	}

	v.recordAssignProperty(valType.Type)
}

// handleAddrOf records pointer usage when taking the address of a non-literal expression of an error type (e.g., `&e`).
func (d dpass) handleAddrOf(n *ast.UnaryExpr) {
	typ, ok := d.TypesInfo.Types[n]
	if !ok {
		return
	}

	tn, ptr, ok := typeutil.TypeNameOf(typ.Type)
	if !ok || !ptr || !d.inCurrentPkg(tn) {
		return
	}

	old, ok := d.ErrorTypes[tn]
	if !ok {
		return
	}

	prop := properties.PointerLiteral
	if old&properties.NonStruct != 0 {
		prop = properties.PointerCast
	}

	d.ErrorTypes[tn] = old | prop
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
	tn, ptr, ok := typeutil.TypeNameOf(tv.Type)
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

			tn, ptr, ok := typeutil.TypeNameOf(tv.Type)
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
	tn, ptr, ok := typeutil.TypeNameOf(typ)
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
	v.handleCompositeLitElts(n)

	litTyp, ok := v.TypesInfo.Types[n]
	if !ok {
		return
	}

	litType := litTyp.Type

	var (
		tn             *types.TypeName
		ptr, namedType bool
	)

	if tn, ptr, namedType = typeutil.TypeNameOf(litType); namedType {
		ptr = ptr || isAddrOf
	} else {
		if p, ok := types.Unalias(litType).(*types.Pointer); ok {
			litType = p.Elem()
		}

		switch t := types.Unalias(litType).(type) {
		case *types.Slice:
			tn, ptr, namedType = typeutil.TypeNameOf(t.Elem())

		case *types.Array:
			tn, ptr, namedType = typeutil.TypeNameOf(t.Elem())

		case *types.Map:
			tn, ptr, namedType = typeutil.TypeNameOf(t.Elem())
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

// handleCompositeLitElts checks elements and fields of a composite literal for conversions to error interfaces.
func (v usageVisitor) handleCompositeLitElts(n *ast.CompositeLit) {
	if len(n.Elts) == 0 {
		return
	}

	litTyp, ok := v.TypesInfo.Types[n]
	if !ok {
		return
	}

	litType := litTyp.Type
	if ptr, ok := types.Unalias(litType).(*types.Pointer); ok {
		litType = ptr.Elem()
	}

	switch t := litType.Underlying().(type) {
	case *types.Slice, *types.Array:
		if elemType := t.(interface{ Elem() types.Type }).Elem(); !typeutil.IsInterfaceWithError(elemType) {
			return
		}

		for _, elt := range n.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				elt = kv.Value
			}

			if valType, ok := v.TypesInfo.Types[elt]; ok {
				v.recordAssignProperty(valType.Type)
			}
		}

	case *types.Map:
		keyIsErr := typeutil.IsInterfaceWithError(t.Key())
		elemIsErr := typeutil.IsInterfaceWithError(t.Elem())

		if !keyIsErr && !elemIsErr {
			return
		}

		for _, elt := range n.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			if keyIsErr {
				if kType, ok := v.TypesInfo.Types[kv.Key]; ok {
					v.recordAssignProperty(kType.Type)
				}
			}

			if elemIsErr {
				if vType, ok := v.TypesInfo.Types[kv.Value]; ok {
					v.recordAssignProperty(vType.Type)
				}
			}
		}

	case *types.Struct:
		for i, elt := range n.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				id, ok := kv.Key.(*ast.Ident)
				if !ok {
					continue
				}

				if obj, ok := v.TypesInfo.Uses[id]; !ok || !typeutil.IsInterfaceWithError(obj.Type()) {
					continue
				}

				if valType, ok := v.TypesInfo.Types[kv.Value]; ok {
					v.recordAssignProperty(valType.Type)
				}
			} else if i < t.NumFields() {
				if !typeutil.IsInterfaceWithError(t.Field(i).Type()) {
					continue
				}

				if valType, ok := v.TypesInfo.Types[elt]; ok {
					v.recordAssignProperty(valType.Type)
				}
			}
		}
	}
}

// handleReturn processes returned values, including multi-valued returns.
func (v usageVisitor) handleReturn(ret *ast.ReturnStmt) {
	v.walkExprs(ret.Results...)

	if v.results == nil || len(ret.Results) == 0 {
		return
	}

	for i := range v.results.Len() {
		if !typeutil.HasErrorMethod(v.results.At(i).Type()) {
			continue
		}

		resType, _ := typeutil.ResultOf(v.TypesInfo, ret.Results, i)
		if resType == nil || resType == types.Typ[types.UntypedNil] || types.IsInterface(resType) {
			continue
		}

		tn, ptr, ok := typeutil.TypeNameOf(resType)
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

// handleErrorWrapper analyzes a function call to determine if it matches patterns like errors.As and identifies the target argument.
// It returns the target argument or nil if the function is not of interest.
func (d dpass) handleErrorWrapper(call *ast.CallExpr) types.Type {
	fun, ok := typeutil.FuncOf(d.TypesInfo, call)
	if !ok {
		return nil // Could not resolve function, might be a func variable.
	}

	wrapper, ok := d.ErrorFuncs[fun.Func]
	if !ok {
		return nil // Not a function we are interested in.
	}

	var (
		targetArgIndex, typeParam int
		address                   bool
	)

	switch wrapper.Type {
	case result.WrapperIs:
		targetArgIndex, typeParam, address = int(wrapper.Target), -1, false

	case result.WrapperAs:
		targetArgIndex, typeParam, address = int(wrapper.Target), -1, true

	case result.WrapperAsType:
		targetArgIndex, typeParam, address = int(wrapper.Source), int(wrapper.Target), true

	case result.FuncAssert:
		targetArgIndex, typeParam, address = -1, int(wrapper.Target), false

	default:
		return nil // Not a function we are interested in.
	}

	if typeParam >= 0 { // Handle generic functions like `errors.AsType[T]`.
		if instance, ok := d.TypesInfo.Instances[fun.Ident]; ok && typeParam < instance.TypeArgs.Len() {
			if typ := instance.TypeArgs.At(typeParam); typeutil.HasErrorMethod(typ) {
				return typ
			}
		}
	}

	if targetArgIndex < 0 {
		return nil
	}

	if fun.MethodExpr {
		// For method expression calls ("(*assert.Assertions).ErrorsAs(a, ...)"),
		// the receiver `a` is the first argument. The argument indices in `errorsAs`
		// are for the function form, so we increment the index to correctly locate
		// the target argument in the method call expression.
		targetArgIndex++
	}

	if len(call.Args) <= targetArgIndex {
		return nil // Maybe called with the result of a multivalued function
	}

	targetArg := call.Args[targetArgIndex]

	typ, ok := d.TypesInfo.Types[targetArg]
	if !ok {
		return nil // errors in type parsing
	}

	if !address {
		return typ.Type
	}

	ptr, ok := typ.Type.Underlying().(*types.Pointer)
	if !ok {
		return nil
	}

	return ptr.Elem()
}

func (v usageVisitor) handleCallExpr(call *ast.CallExpr) {
	if target := v.handleErrorWrapper(call); target != nil {
		tn, ptr, ok := typeutil.TypeNameOf(target)
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

	if typeutil.BuiltinNew(v.TypesInfo, call) {
		v.handleBuiltinNew(call.Args[0])

		return
	}

	// not an error wrapper
	ast.Walk(v, call.Fun)
	v.walkExprs(call.Args...)
	v.handleCallArgs(call)
}

func (v usageVisitor) handleBuiltinNew(arg ast.Expr) {
	switch e := ast.Unparen(arg).(type) {
	case *ast.CompositeLit:
		// treat new(MyError{...}) like &MyError{...}.
		v.handleCompositeLit(e, true)

	case *ast.CallExpr:
		if tv, ok := v.TypesInfo.Types[e.Fun]; ok && tv.IsType() {
			v.walkExprs(e.Args...)

			// treat new(myIntError(1)) like (*myIntError)(&1).
			if tn, ptr, ok := typeutil.TypeNameOf(tv.Type); ok && !ptr {
				v.addTypePropertyInCurrentPackage(tn, properties.PointerCast)
			}

			return
		}

		ast.Walk(v, arg)

	default:
		ast.Walk(v, arg)
	}
}

func (v usageVisitor) handleCallArgs(call *ast.CallExpr) {
	if len(call.Args) == 0 {
		return
	}

	funType, ok := v.TypesInfo.Types[call.Fun]
	if !ok {
		return
	}

	sig, ok := funType.Type.Underlying().(*types.Signature)
	if !ok || sig.Params().Len() == 0 {
		return
	}

	params := sig.Params()
	numParams := params.Len()

	// Multi-valued call: fn(g()) where g() returns a tuple.
	if len(call.Args) == 1 && numParams > 1 {
		for i := range numParams {
			if !typeutil.IsInterfaceWithError(params.At(i).Type()) {
				continue
			}

			if argType, _ := typeutil.ResultOf(v.TypesInfo, call.Args, i); argType != nil {
				v.recordAssignProperty(argType)
			}
		}

		return
	}

	last := numParams - 1

	for i, arg := range call.Args {
		var paramType types.Type

		switch {
		case sig.Variadic() && i >= last:
			if call.Ellipsis.IsValid() {
				continue
			}

			slice, ok := params.At(last).Type().(*types.Slice)
			if !ok {
				continue
			}

			paramType = slice.Elem()

		case i < numParams:
			paramType = params.At(i).Type()

		default:
			return
		}

		if !typeutil.IsInterfaceWithError(paramType) {
			continue
		}

		if argType, ok := v.TypesInfo.Types[arg]; ok {
			v.recordAssignProperty(argType.Type)
		}
	}
}

// recordAssignProperty analyzes the given type to determine if it's a pointer or
// value error and records a [properties.PointerAssign] or [properties.ValueAssign] property respectively.
func (d dpass) recordAssignProperty(typ types.Type) {
	if typ == nil || typ == types.Typ[types.UntypedNil] || types.IsInterface(typ) {
		return
	}

	tn, ptr, ok := typeutil.TypeNameOf(typ)
	if !ok {
		return
	}

	prop := properties.ValueAssign
	if ptr {
		prop = properties.PointerAssign
	}

	d.addTypePropertyInCurrentPackage(tn, prop)
}

// addTypePropertyInCurrentPackage sets a property on a type if it's a known error type
// in the current package and the property isn't yet set.
func (d dpass) addTypePropertyInCurrentPackage(tn *types.TypeName, property properties.ErrorProperty) {
	if !d.inCurrentPkg(tn) {
		return // Only relevant for types defined in the current package
	}

	if old, ok := d.ErrorTypes[tn]; ok && old&property != property { // known, but property isn't set.
		d.ErrorTypes[tn] |= property
	}
}
