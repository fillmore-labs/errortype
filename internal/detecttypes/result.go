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

	"fillmore-labs.com/errortype/internal/detecttypes/result"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// createResult combines all determined type information into the final analyzer result.
// It merges types from the current package and dependencies (facts) and local overrides,
// with local overrides having the highest precedence.
func (d dpass) createResult(ctx context.Context, eTypes result.ErrorTypes) result.Result {
	defer trace.StartRegion(ctx, "result").End()

	for tn, prop := range d.ErrorTypes {
		errorType := prop.DeterminedType()

		if d.inCurrentPkg(tn) {
			// Export this information as a fact when the type is defined in the current package.
			// These facts can then be consumed by analyzers running on packages dependent on this one.
			d.ExportObjectFact(tn, errorType.Fact())
		}

		// Add type to the result.
		// Local overrides will overwrite any existing entries from facts, only for this package.
		eTypes[tn] = errorType
	}

	// Convert map to slice for the result.
	return result.New(eTypes)
}

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

// logResults logs the properties and the determined error type for each type in the PropertyMap.
func (d dpass) logResults(_ context.Context, logger *log.Logger) {
	pkgPath := typeutil.PkgPath(d.Fset, d.Pkg)
	qf := types.RelativeTo(d.Pkg)

	for tn, errortype := range d.ErrorTypes.AllSorted {
		typeName := types.TypeString(tn.Type(), qf)
		determinedType := errortype.DeterminedType()

		logger.Printf("%s %s: %s (%s)\n", pkgPath, typeName, determinedType, errortype)
	}
}
