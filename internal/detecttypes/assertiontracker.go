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
	"go/ast"
	"go/types"

	"fillmore-labs.com/errortype/internal/astutil"
	"fillmore-labs.com/errortype/internal/detecttypes/properties"
	"fillmore-labs.com/errortype/internal/detecttypes/result"
	"fillmore-labs.com/errortype/internal/typeutil"
)

type assertionTracker struct {
	dpass
	local     *localAssertions
	firstDecl map[*types.TypeName]declRef
	firstDir  map[*types.TypeName]dirRef
}

// findErrorAssertions checks for error assertion declarations (`var _ error = ...`).
// It also handles sentinel errors with explicit types, e.g., `var ErrSomething error = ...`.
func (t *assertionTracker) findErrorAssertions(decl *ast.GenDecl, varDecl *ast.ValueSpec, eTypes result.ErrorTypes, heuristicVar bool) {
	var (
		dir    ast.Directive
		hasDir bool
		parsed bool
	)

	for i, id := range varDecl.Names {
		typ, expr := typeutil.ResultOf(t.TypesInfo, varDecl.Values, i)
		if typ == nil {
			t.LogErrorf(varDecl, "can't get value for variable %q", id.Name)

			continue
		}

		if id.Name != "_" {
			if heuristicVar {
				t.recordVarProperty(typ, properties.PointerVar, properties.ValueVar)
			}

			continue
		}

		// A compile-time assertion only declares intent.
		if types.IsInterface(typ) {
			continue
		}

		tn, ptr, ok := typeutil.CanonicalTypeNameOf(typ)
		if !ok || !typeutil.PackageLevel(tn) {
			continue // struct { embedded }, nil, or non-package-level type
		}

		if t.inCurrentPkg(tn) {
			if heuristicVar {
				t.addTypePropertyInCurrentPackage(tn, declProperty(ptr))
				t.recordPlainAssertion(tn, expr, ptr)
			}

			continue
		}

		if !parsed {
			dir, hasDir = astutil.Directive(decl, varDecl.Doc, varDecl.Comment, "errortype")
			parsed = true
		}

		if hasDir {
			t.handleExternalDirective(tn, expr, ptr, dir.Name)

			continue
		}

		if heuristicVar && eTypes[tn]&result.ExpectedMask == result.Undecided {
			t.addLocalProperty(tn, declProperty(ptr))
			t.recordPlainAssertion(tn, expr, ptr)
		}
	}
}

func (t *assertionTracker) addLocalProperty(tn *types.TypeName, prop properties.ErrorProperty) {
	if t.local == nil {
		t.local = &localAssertions{types: make(properties.Properties)}
	} else if t.local.types == nil {
		t.local.types = make(properties.Properties)
	}

	t.local.types[tn] |= prop
}

func declProperty(ptr bool) properties.ErrorProperty {
	if ptr {
		return properties.PointerDecl
	}

	return properties.ValueDecl
}

func (t *assertionTracker) handleExternalDirective(tn *types.TypeName, expr ast.Expr, ptr bool, dirName string) {
	var dirType result.ErrorType
	if err := dirType.UnmarshalText([]byte(dirName)); err != nil {
		t.addIssue(result.UnknownDirective{
			Expr:      expr,
			Directive: dirName,
		})

		return
	}

	if dirType == result.Pointer && !ptr || dirType == result.Value && ptr {
		t.addIssue(result.MismatchedDirective{
			Expr:      expr,
			TypeName:  tn,
			Directive: dirType,
		})

		return
	}

	t.addLocalProperty(tn, directiveOverrideProperty(dirType))

	if t.firstDir == nil {
		t.firstDir = make(map[*types.TypeName]dirRef)
	}

	if first, exists := t.firstDir[tn]; !exists {
		t.firstDir[tn] = dirRef{expr: expr, dir: dirType}
	} else if dirType != first.dir {
		t.addIssue(result.ConflictingDirective{
			Expr:       expr,
			TypeName:   tn,
			Directive:  dirType,
			Related:    first.expr,
			RelatedDir: first.dir,
		})
	}
}

func (t *assertionTracker) recordPlainAssertion(tn *types.TypeName, expr ast.Expr, ptr bool) {
	if t.firstDecl == nil {
		t.firstDecl = make(map[*types.TypeName]declRef)
	}

	first, exists := t.firstDecl[tn]
	if !exists {
		t.firstDecl[tn] = declRef{expr: expr, ptr: ptr}

		return
	}

	if !first.reported && ptr != first.ptr {
		first.reported = true
		t.firstDecl[tn] = first
		t.addIssue(result.ContradictoryAssertion{
			Expr:     expr,
			TypeName: tn,
			Related:  first.expr,
			FirstPtr: first.ptr,
		})
	}
}

func (t *assertionTracker) addIssue(issue result.AssertionIssue) {
	if t.local == nil {
		t.local = &localAssertions{}
	}

	t.local.issues = append(t.local.issues, issue)
}

func directiveOverrideProperty(dirType result.ErrorType) properties.ErrorProperty {
	switch dirType {
	case result.Pointer:
		return properties.PointerOverride

	case result.Value:
		return properties.ValueOverride

	case result.Suppress:
		return properties.SuppressOverride

	default:
		return 0
	}
}
