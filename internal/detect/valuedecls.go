// Copyright 2025-2026 Oliver Eikemeier. All Rights Reserved.
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
	"go/types"
	"runtime/trace"
	"strings"

	"fillmore-labs.com/errortype/internal/detect/properties"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// processValueSpecs processes variable and constant declarations to infer error
// type properties. It identifies whether an error type should be used as a pointer
// or a value based on two patterns:
//
//  1. Error Assertions: A declaration like `var _ error = &T{}` marks `T`
//     as a pointer error, while `var _ error = T{}` marks it as a value error.
//
//  2. Sentinel Errors: A declaration like `var ErrSomething = &T{}` indicates
//     a pointer error, while `var ErrSomething = T{}` indicates a value error.
//
//     It also detects sentinel constant, e.g. `const ErrFoo = stringError("foo")`,
//     although this should be rare.
//
// Discovered properties are recorded for analysis. If the type `T` is defined
// in the current package, this property is exported as a fact.
//
// If `T` is from an external package, the property is checked for consistency.
func (p pass) processValueSpecs(ctx context.Context) {
	defer trace.StartRegion(ctx, "valueSpecs").End()

	for varDecl := range typeutil.AllValueDecls(p.Files) {
		// Handle sentinel errors, e.g., `var ErrSomething = ...` where the type is inferred.
		if varDecl.Type == nil {
			p.findSentinelErrorValues(varDecl)

			continue
		}

		// Handle error assertions, e.g., `var _ error = ...` where the type is explicit.
		// It also handles sentinel errors with explicit types, e.g., `var ErrSomething error = ...`.
		p.findErrorAssertions(varDecl)
	}
}

// findSentinelErrorValues checks for sentinel error declarations (`var Err... = ...`),
// including multi-return assignments like `var ErrFoo, _ = newError(...)`.
func (p pass) findSentinelErrorValues(varDecl *ast.ValueSpec) {
	var typeAt func(int) types.Type

	switch len(varDecl.Values) {
	case 0:
		return // No values

	case len(varDecl.Names):
		typeAt = func(i int) types.Type {
			return p.TypesInfo.Types[varDecl.Values[i]].Type
		}

	case 1:
		// Multiple names with a single value: must be a multi-return function call.
		tup, ok := p.TypesInfo.Types[varDecl.Values[0]].Type.(*types.Tuple)
		if !ok || tup.Len() != len(varDecl.Names) {
			return // should not happen
		}

		typeAt = func(i int) types.Type {
			return tup.At(i).Type()
		}

	default:
		return // should not happen
	}

	for i, id := range varDecl.Names {
		if !strings.HasPrefix(id.Name, "Err") && !strings.HasPrefix(id.Name, "err") {
			continue
		}

		typ := typeAt(i)
		if typ == nil || !typeutil.HasErrorMethod(typ) {
			continue // Not an error type
		}

		p.recordVarProperty(typ)
	}
}

// findErrorAssertions checks for error assertion declarations (`var _ error = ...`).
// It also handles sentinel errors with explicit types, e.g., `var ErrSomething error = ...`.
func (p pass) findErrorAssertions(varDecl *ast.ValueSpec) {
	if tv, ok := p.TypesInfo.Types[varDecl.Type]; !ok || !typeutil.HasErrorMethod(tv.Type) {
		return
	}

	if len(varDecl.Values) > len(varDecl.Names) { // should not happen
		return
	}

	for i, value := range varDecl.Values {
		tv, ok := p.TypesInfo.Types[value]
		if !ok || !tv.IsValue() { // should not happen
			name := varDecl.Names[i].Name
			p.LogErrorf(value, "can't get value for variable %s", name)

			continue
		}

		p.recordVarProperty(tv.Type)
	}
}

// recordVarProperty analyzes the given type to determine if it's a pointer or
// value error and records a [PointerVar] or [ValueVar] property respectively.
func (p pass) recordVarProperty(typ types.Type) {
	// Interfaces are not concrete error types.
	if types.IsInterface(typ) {
		return
	}

	// We need a named type to associate the property with. This skips anonymous structs and other unnamed types.
	tn, ptr, ok := typeutil.TypeNameOf(typ)
	if !ok {
		return // struct { embedded } or nil
	}

	errortype := properties.ValueVar
	if ptr {
		errortype = properties.PointerVar
	}

	// Record usage in the property map.
	// If the type is defined in the current package, it determines usage.
	// Otherwise, it's ignored.
	p.addTypePropertyInCurrentPackage(tn, errortype)
}
