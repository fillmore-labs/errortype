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
	"go/types"
	"runtime/trace"
	"strings"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// processVarSpecs processes variable declarations to infer error type properties.
// It identifies whether an error type should be used as a pointer or a value
// based on two patterns:
//
//  1. Error Assertions: A declaration like `var _ error = &T{}` marks `T`
//     as a pointer error, while `var _ error = T{}` marks it as a value error.
//
//  2. Sentinel Errors: A declaration like `var ErrSomething = &T{}` indicates
//     a pointer error, while `var ErrSomething = T{}` indicates a value error.
//
// Discovered properties are recorded for analysis. If the type `T` is defined
// in the current package, this property is exported as a fact.
//
// If `T` is from an external package, the property is checked for consistency.
func (p pass) processVarSpecs(ctx context.Context) {
	defer trace.StartRegion(ctx, "varSpecs").End()

	for varDecl := range p.AllVarDecls {
		// Handle sentinel errors, e.g., `var ErrSomething = ...` where the type is inferred.
		if varDecl.Type == nil {
			p.findSentinelErrors(varDecl)

			continue
		}

		// Handle error assertions, e.g., `var _ error = ...` where the type is explicit.
		p.findErrorAssertions(varDecl)
	}
}

// findSentinelErrors checks for sentinel error declarations (`var Err...`).
func (p pass) findSentinelErrors(varDecl *ast.ValueSpec) {
	for i, id := range varDecl.Names {
		if len(varDecl.Values) <= i {
			break
		}

		if !strings.HasPrefix(id.Name, "Err") && !strings.HasPrefix(id.Name, "err") {
			continue
		}

		value := varDecl.Values[i]

		tv, ok := p.TypesInfo.Types[value]
		if !ok || !typeutil.HasErrorMethod(tv.Type) {
			continue // Not an error type
		}

		p.recordVarProperty(tv.Type)
	}
}

// findErrorAssertions checks for error assertion declarations (`var _ error = ...`).
func (p pass) findErrorAssertions(varDecl *ast.ValueSpec) {
	if tv, ok := p.TypesInfo.Types[varDecl.Type]; !ok || !typeutil.HasErrorMethod(tv.Type) {
		return
	}

	for i, value := range varDecl.Values {
		tv, ok := p.TypesInfo.Types[value]
		if !ok || !tv.IsValue() { // should not happen
			name := "<unknown>"
			if len(varDecl.Names) > i {
				name = varDecl.Names[i].Name
			}

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

	errortype := ValueVar
	if ptr {
		errortype = PointerVar
	}

	// Record usage in the property map.
	// If the type is defined in the current package, it determines usage.
	// Otherwise, it's ignored.
	p.addTypePropertyInCurrentPackage(tn, errortype)
}
