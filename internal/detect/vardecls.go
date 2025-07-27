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
// in the current package, this property is exported as a fact. If `T` is from
// an external package, the property is treated as a local override.
func (p pass) processVarSpecs() {
	for varspec := range p.AllVarDecls {
		// Handle sentinel errors, e.g., `var ErrSomething = ...` where the type is inferred.
		if varspec.Type == nil {
			p.findSentinelErrors(varspec)

			continue
		}

		// Handle error assertions, e.g., `var _ error = ...` where the type is explicit.
		p.findErrorAssertions(varspec)
	}
}

// findSentinelErrors checks for sentinel error declarations (`var Err...`).
func (p pass) findSentinelErrors(varspec *ast.ValueSpec) {
	for i, id := range varspec.Names {
		if len(varspec.Values) <= i {
			break
		}

		if !strings.HasPrefix(id.Name, "Err") && !strings.HasPrefix(id.Name, "err") {
			continue
		}

		value := varspec.Values[i]

		tv, ok := p.TypesInfo.Types[value]
		if !ok || !typeutil.HasErrorMethod(tv.Type) {
			continue
		}

		p.recordErrorProperty(tv.Type)
	}
}

// findErrorAssertions checks for error assertion declarations (`var _ error = ...`).
func (p pass) findErrorAssertions(varspec *ast.ValueSpec) {
	if tv, ok := p.TypesInfo.Types[varspec.Type]; !ok || !typeutil.HasErrorMethod(tv.Type) {
		return
	}

	for i, value := range varspec.Values {
		tv, ok := p.TypesInfo.Types[value]
		if !ok || !tv.IsValue() { // should not happen
			var name string
			if len(varspec.Names) > i {
				name = varspec.Names[i].Name
			}

			p.LogErrorf(value, "can't get type from value %s", name)

			continue
		}

		p.recordErrorProperty(tv.Type)
	}
}

// recordErrorProperty analyzes the given type to determine if it's a pointer or
// value error and records the property.
func (p pass) recordErrorProperty(typ types.Type) {
	// Interfaces are not concrete error types.
	if types.IsInterface(typ) {
		return
	}

	// We need a named type to associate the property with. This skips anonymous structs and other unnamed types.
	tn, isPtr, ok := typeutil.TypeNameOf(typ)
	if !ok {
		return // struct { embedded } or nil
	}

	errortype := ValueVar
	if isPtr {
		errortype = PointerVar
	}

	// Record usage in the property map.
	// If the type is defined in the current package, it determines usage.
	// Otherwise, it's a local override.
	p.AddTypeProperty(tn, errortype)
}
