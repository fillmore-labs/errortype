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
	"go/types"
	"log"
	"runtime/trace"
	"slices"

	"fillmore-labs.com/errortype/internal/detecttypes/result"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// createResult exports facts for current-package error types and builds the final
// analyzer result by combining imported facts, current-package types, package-local
// overrides, and remaining assertion issues.
func (d dpass) createResult(ctx context.Context, eTypes result.ErrorTypes, local *localAssertions) result.Result {
	defer trace.StartRegion(ctx, "result").End()

	for tn, prop := range d.ErrorTypes {
		errorType := prop.DeterminedType()

		if d.inCurrentPkg(tn) {
			// Export this information as a fact when the type is defined in the current package.
			// These facts can then be consumed by analyzers running on packages dependent on this one.
			d.ExportObjectFact(tn, errorType.Fact())
		}

		// Record the current package's determined type in the result.
		eTypes[tn] = errorType
	}

	issues := applyLocalAssertions(eTypes, local)

	return result.New(eTypes, issues)
}

// applyLocalAssertions merges resolved package-local type assertions and directives
// into eTypes (for the current package only) and returns the remaining assertion issues.
//
// Local overrides are merged before filtering issues so that a contradictory plain
// assertion resolved by a directive or config override is dropped alongside those
// resolved by other heuristics. Directive issues are always preserved.
// Unresolved entries in local.types are skipped so they do not overwrite imported facts.
// It filters local.issues in place.
func applyLocalAssertions(eTypes result.ErrorTypes, local *localAssertions) []result.AssertionIssue {
	if local == nil {
		return nil
	}

	for tn, prop := range local.types {
		if errorType := prop.DeterminedType(); errorType != result.Undecided {
			// Local assertions and directives override imported facts only for this package.
			eTypes[tn] = errorType
		}
	}

	if len(local.issues) == 0 {
		return nil
	}

	local.issues = slices.DeleteFunc(local.issues, func(issue result.AssertionIssue) bool {
		c, ok := issue.(result.ContradictoryAssertion)

		return ok && eTypes[c.TypeName]&result.ExpectedMask != result.Undecided
	})

	return local.issues
}

// extractFacts collects error type facts exported by dependency packages.
func (d dpass) extractFacts() result.ErrorTypes {
	eTypes := make(result.ErrorTypes)

	for _, fact := range d.AllObjectFacts() {
		f, ok := fact.Fact.(*result.ErrorType)
		if !ok {
			continue
		}

		tn, ok := fact.Object.(*types.TypeName)
		if !ok {
			continue
		}

		eTypes[tn] = *f
	}

	return eTypes
}

// logResults logs the properties and determined error type for each current-package
// error type and any package-local overrides.
func (d dpass) logResults(_ context.Context, logger *log.Logger, local *localAssertions) {
	pkgPath := typeutil.PkgPath(d.Fset, d.Pkg, d.Files)
	qf := types.RelativeTo(d.Pkg)

	for tn, prop := range d.ErrorTypes.AllSorted {
		typeName := types.TypeString(tn.Type(), qf)
		errorType := prop.DeterminedType()

		logger.Printf("%s %s: %s (%s)\n", pkgPath, typeName, errorType, prop)
	}

	if local != nil {
		for tn, prop := range local.types.AllSorted {
			typeName := types.TypeString(tn.Type(), qf)
			errorType := prop.DeterminedType()

			logger.Printf("%s local %s: %s (%s)\n", pkgPath, typeName, errorType, prop)
		}
	}
}
