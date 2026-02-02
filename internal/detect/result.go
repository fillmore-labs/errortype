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
	"go/types"
	"iter"
	"log"
	"maps"
	"runtime/trace"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/facts"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// createResult combines all determined type information into the final analyzer result.
// It merges types from the current package and dependencies (facts) and local overrides,
// with local overrides having the highest precedence.
func (p pass) createResult(ctx context.Context) facts.Result {
	defer trace.StartRegion(ctx, "result").End()

	// Add types from dependencies (via facts).
	allFacts := p.AllObjectFacts()
	determinedTypes := make(map[*types.TypeName]facts.ErrorFact, len(allFacts))
	maps.Insert(determinedTypes, extractErrorTypes(allFacts))

	for tn, prop := range p.DetectedTypes {
		errorType := prop.DeterminedType()

		if p.inCurrentPkg(tn) {
			// Export this information as a fact when the type is defined in the current package.
			// These facts can then be consumed by analyzers running on packages dependent on this one.
			p.ExportObjectFact(tn, &errorType)
		}

		// Add type to the result.
		// Local overrides will overwrite any existing entries from facts, only for this package.
		determinedTypes[tn] = errorType
	}

	// Convert map to slice for the result.
	return createResult(determinedTypes)
}

func createResult(determinedTypes map[*types.TypeName]facts.ErrorFact) facts.Result {
	typs := make([]facts.ResultInfo, 0, len(determinedTypes))
	for tn, errorType := range determinedTypes {
		typs = append(typs, facts.ResultInfo{TypeName: tn, ErrorType: errorType})
	}

	return facts.Result{Types: typs}
}

// extractErrorTypes processes imported [analysis.ObjectFact]s and returns a sequence of *types.TypeName with errortypes.ErrorType.
func extractErrorTypes(allFacts []analysis.ObjectFact) iter.Seq2[*types.TypeName, facts.ErrorFact] {
	return func(yield func(*types.TypeName, facts.ErrorFact) bool) {
		for _, f := range allFacts {
			fact, ok := f.Fact.(*facts.ErrorFact)
			if !ok || fact == nil {
				continue
			}

			if tn, ok := f.Object.(*types.TypeName); ok && !yield(tn, *fact) {
				return
			}
		}
	}
}

// logResults logs the properties and the determined error type for each type in the PropertyMap.
func (p pass) logResults(_ context.Context, logger *log.Logger) {
	pkgPath := typeutil.PkgPath(p.Pass)
	qf := types.RelativeTo(p.Pkg)

	for tn, errortype := range p.DetectedTypes.AllSorted {
		typeName := types.TypeString(tn.Type(), qf)
		determinedType := errortype.DeterminedType()

		logger.Printf("%s %s: %s (%s)", pkgPath, typeName, determinedType, errortype)
	}
}
