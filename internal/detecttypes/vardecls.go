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

package detecttypes

import (
	"context"
	"go/ast"
	"go/token"
	"go/types"
	"runtime/trace"
	"slices"
	"strings"

	"fillmore-labs.com/errortype/internal/detecttypes/properties"
	"fillmore-labs.com/errortype/internal/detecttypes/result"
	"fillmore-labs.com/errortype/internal/errorlike"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// localAssertions holds package-local state collected from compile-time
// `var _ error` assertions: override and tie-breaker properties for imported
// types (never exported as facts) and assertion configuration or contradiction issues.
// A nil *localAssertions indicates that no local overrides or issues were recorded.
type localAssertions struct {
	types  properties.Properties
	issues []result.AssertionIssue
}

type declRef struct {
	expr     ast.Expr
	ptr      bool
	reported bool
}

type dirRef struct {
	expr ast.Expr
	dir  result.ErrorType
}

// processVarDecls processes variable and constant declarations to infer error
// type properties and record assertion configuration issues. It identifies whether
// an error type should be used as a pointer or a value based on two patterns:
//
//  1. Error Assertions: A declaration like `var _ error = (*T)(nil)` marks `T`
//     as a pointer error, while `var _ error = T{}` marks it as a value error.
//     Since they only declare intent, they take precedence over sentinel errors.
//     For imported types, plain assertions act as local tie-breakers when upstream
//     is undecided; assertions tagged with `//errortype:pointer`, `//errortype:value`,
//     or `//errortype:suppress` act as local overrides.
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
// If `T` is from an external package, any local assertion or directive override applies
// only to the current package.
func (d dpass) processVarDecls(ctx context.Context, eTypes result.ErrorTypes, heuristicVar bool) *localAssertions {
	defer trace.StartRegion(ctx, "valueSpecs").End()

	hasUndetermined := heuristicVar && d.ErrorTypes.HasUndeterminedErrors()
	tracker := assertionTracker{dpass: d}

	for decl, varDecl := range typeutil.AllVarDecls(d.Files) {
		// Handle sentinel errors, e.g., `var ErrSomething = ...` where the type is inferred.
		if varDecl.Type == nil {
			if hasUndetermined {
				d.findSentinelErrorValues(varDecl)
			}

			continue
		}

		if len(varDecl.Values) == 0 {
			continue
		}

		if !hasUndetermined && (decl.Tok != token.VAR || !hasBlankName(varDecl.Names)) {
			continue
		}

		if tv, ok := d.TypesInfo.Types[varDecl.Type]; !ok || !errorlike.Implements(tv.Type) {
			continue
		}

		// Handle error assertions, e.g., `var _ error = ...` where the type is explicit.
		// It also handles sentinel errors with explicit types, e.g., `var ErrSomething error = ...`.
		tracker.findErrorAssertions(decl, varDecl, eTypes, heuristicVar)
	}

	return tracker.local
}

func hasBlankName(names []*ast.Ident) bool {
	return slices.ContainsFunc(names, func(id *ast.Ident) bool { return id.Name == "_" })
}

// findSentinelErrorValues checks for sentinel error declarations (`var Err... = ...`).
func (d dpass) findSentinelErrorValues(varDecl *ast.ValueSpec) {
	for i, id := range varDecl.Names {
		if !hasErrPrefix(id.Name) {
			continue
		}

		typ, _ := typeutil.ResultOf(d.TypesInfo, varDecl.Values, i)
		if typ == nil || !errorlike.Implements(typ) {
			continue // Not an error type
		}

		d.recordVarProperty(typ, properties.PointerVar, properties.ValueVar)
	}
}

func hasErrPrefix(name string) bool {
	return strings.HasPrefix(name, "Err") || strings.HasPrefix(name, "err")
}

// recordVarProperty analyzes the given type to determine if it's a pointer or
// value error and records pointerProp or valueProp respectively.
func (d dpass) recordVarProperty(typ types.Type, pointerProp, valueProp properties.ErrorProperty) {
	// Interfaces are not concrete error types.
	if types.IsInterface(typ) {
		return
	}

	// We need a named type to associate the property with. This skips anonymous structs and other unnamed types.
	tn, ptr, ok := typeutil.CanonicalTypeNameOf(typ)
	if !ok {
		return // struct { embedded } or nil
	}

	errortype := valueProp
	if ptr {
		errortype = pointerProp
	}

	// Record usage in the property map.
	// If the type is defined in the current package, it determines usage.
	// Otherwise, it's ignored.
	d.addTypePropertyInCurrentPackage(tn, errortype)
}
