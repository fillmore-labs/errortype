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
	"go/types"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// processUsage processes all function declarations in the current package,
// visiting their bodies to perform error usage analysis.
func (p pass) processUsage() {
	u := usageVisitor{pass: p}

	for f := range p.AllFuncDecls {
		if f.Body == nil {
			continue
		}

		u.lastResult = typeutil.HasErrorResult(p.TypesInfo, f.Type.Results)

		ast.Walk(u, f.Body)
	}
}

type usageVisitor struct {
	pass
	lastResult int
}

func (v usageVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.AssignStmt:
		// Analyze the right-hand side of `:=` and `=` assignments.
		v.walkExprs(n.Rhs)

		return nil // Expressions handled, stop descending.

	case *ast.ValueSpec:
		// Analyze initial values in `var` declarations.
		v.walkExprs(n.Values)

		return nil // Expressions handled, stop descending.

	case *ast.ReturnStmt:
		v.handleReturn(n)

		return nil // Expressions handled, stop descending.

	case *ast.SendStmt:
		// Analyze the value sent to a channel.
		a := assignVisitor{pass: v.pass}
		ast.Walk(a, n.Value)

		return nil // Expression handled, stop descending.

	case *ast.FuncLit:
		// A function literal defines a new function.
		// We inspect its body for how it returns error types.
		u := usageVisitor{
			pass:       v.pass,
			lastResult: typeutil.HasErrorResult(v.TypesInfo, n.Type.Results),
		}

		ast.Walk(u, n.Body)

		return nil // Expression handled, stop descending.

	case *ast.CallExpr:
		v.handleCallExpr(n)

		return nil // Expressions handled, stop descending.

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

// handleReturn processes returned values, T{} or &T{}.
func (v usageVisitor) handleReturn(ret *ast.ReturnStmt) {
	v.walkExprs(ret.Results)

	if v.lastResult < 0 || len(ret.Results) <= v.lastResult {
		return
	}

	res := ret.Results[v.lastResult]

	resType := v.TypesInfo.Types[res]
	if !resType.IsValue() { // should not happen
		v.LogErrorf(res, "Expected value, got %#v", resType)
	}

	if resType.IsNil() {
		return // nil is fine.
	}

	tn, isPtr, ok := typeutil.TypeNameOf(resType.Type)
	if !ok {
		return // Not a named type.
	}

	property := ValueReturn
	if isPtr {
		property = PointerReturn
	}

	v.addTypePropertyInCurrentPackage(tn, property)
}

func (p pass) handleCallExpr(n *ast.CallExpr) {
	_, _, targetArgIndex := typeutil.IsErrorAs(p.TypesInfo, n)
	if targetArgIndex < 0 { // not an errors.As-like function
		p.walkExprs(n.Args)

		if f, ok := n.Fun.(*ast.FuncLit); ok { // For immediately invoked function literals, examine their body.
			u := usageVisitor{
				pass:       p,
				lastResult: typeutil.HasErrorResult(p.TypesInfo, f.Type.Results),
			}

			ast.Walk(u, f.Body)
		}

		return
	}

	if len(n.Args) <= targetArgIndex {
		return // Maybe called with the result of a multivalued function
	}

	targetArg := n.Args[targetArgIndex]

	typ, ok := p.TypesInfo.Types[targetArg]
	if !ok {
		return
	}

	ptr, ok := typ.Type.Underlying().(*types.Pointer)
	if !ok {
		return
	}

	tn, isPtr, ok := typeutil.TypeNameOf(ptr.Elem())
	if !ok {
		return // Not a named type.
	}

	property := ValueTarget
	if isPtr {
		property = PointerTarget
	}

	p.addTypePropertyInCurrentPackage(tn, property)
}

// walkExprs applies the assignVisitor to each expression in the given list.
// This is used to analyze expressions in assignment, return, or declaration contexts.
func (p pass) walkExprs(exprs []ast.Expr) {
	a := assignVisitor{pass: p}
	for _, expr := range exprs {
		ast.Walk(a, expr)
	}
}

// addTypePropertyInCurrentPackage sets a property on a type if it's a known error type
// in the current package and the property isn't yet set.
func (p pass) addTypePropertyInCurrentPackage(tn *types.TypeName, property ErrorProperty) {
	if tn.Pkg() != p.Pkg {
		return // Only relevant for types defined in the current package
	}

	old, ok := p.GetTypeProperty(tn)
	if !ok {
		return // Not a known error type
	}

	if old&property == 0 { // property isn't set.
		p.SetTypeProperty(tn, old|property)
	}
}
