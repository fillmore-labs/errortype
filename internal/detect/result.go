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
	"maps"
	"runtime/trace"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/errortypes"
)

// createResult combines all determined type information into the final analyzer result.
// It merges types from the current package and dependencies (facts) and local overrides,
// with local overrides having the highest precedence.
func (p pass) createResult(ctx context.Context) errortypes.Result {
	defer trace.StartRegion(ctx, "result").End()

	// Add types from dependencies (via facts).
	facts := p.AllObjectFacts()
	determinedTypes := make(map[*types.TypeName]errortypes.ErrorType, len(facts))
	maps.Insert(determinedTypes, extractErrorTypes(facts))

	for tn, prop := range p.PropertyMap {
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

func createResult(determinedTypes map[*types.TypeName]errortypes.ErrorType) errortypes.Result {
	typs := make([]errortypes.ResultInfo, 0, len(determinedTypes))
	for tn, errorType := range determinedTypes {
		typs = append(typs, errortypes.ResultInfo{TypeName: tn, ErrorType: errorType})
	}

	return errortypes.Result{Types: typs}
}

// extractErrorTypes processes imported [analysis.ObjectFact]s and returns a sequence of *types.TypeName with errortypes.ErrorType.
func extractErrorTypes(facts []analysis.ObjectFact) iter.Seq2[*types.TypeName, errortypes.ErrorType] {
	return func(yield func(*types.TypeName, errortypes.ErrorType) bool) {
		for _, f := range facts {
			fact, ok := f.Fact.(*errortypes.ErrorType)
			if !ok || fact == nil {
				continue
			}

			tn, ok := f.Object.(*types.TypeName)
			if !ok {
				continue
			}

			if !yield(tn, *fact) {
				return
			}
		}
	}
}
