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

// processVarDecls processes variable and constant declarations to infer error
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
func (p pass) processVarDecls(ctx context.Context) {
	defer trace.StartRegion(ctx, "valueSpecs").End()

	for varDecl := range typeutil.AllVarDecls(p.Files) {
		// Handle sentinel errors, e.g., `var ErrSomething = ...` where the type is inferred.
		if varDecl.Type == nil {
			p.findSentinelErrorValues(varDecl)

			continue
		}

		if len(varDecl.Values) == 0 {
			continue
		}

		if tv, ok := p.TypesInfo.Types[varDecl.Type]; !ok || !typeutil.HasErrorMethod(tv.Type) {
			continue
		}

		// Handle error assertions, e.g., `var _ error = ...` where the type is explicit.
		// It also handles sentinel errors with explicit types, e.g., `var ErrSomething error = ...`.
		p.findErrorAssertions(varDecl)
	}
}

// findSentinelErrorValues checks for sentinel error declarations (`var Err... = ...`).
func (p pass) findSentinelErrorValues(varDecl *ast.ValueSpec) {
	for i, id := range varDecl.Names {
		if !hasErrPrefix(id.Name) {
			continue
		}

		typ, _ := typeutil.ResultOf(p.TypesInfo, varDecl.Values, i)
		if typ == nil || !typeutil.HasErrorMethod(typ) {
			continue // Not an error type
		}

		p.recordVarProperty(typ)
	}
}

func hasErrPrefix(name string) bool {
	return strings.HasPrefix(name, "Err") || strings.HasPrefix(name, "err")
}

// findErrorAssertions checks for error assertion declarations (`var _ error = ...`).
// It also handles sentinel errors with explicit types, e.g., `var ErrSomething error = ...`.
func (p pass) findErrorAssertions(varDecl *ast.ValueSpec) {
	for i := range varDecl.Names {
		typ, _ := typeutil.ResultOf(p.TypesInfo, varDecl.Values, i)
		if typ == nil {
			name := varDecl.Names[i].Name
			p.LogErrorf(varDecl, "can't get value for variable %q", name)

			continue
		}

		p.recordVarProperty(typ)
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
