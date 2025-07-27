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
	"go/types"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/errortypes"
)

// ResultInfo holds the determined pointer-ness for a type,
// identified by its *types.TypeName.
type ResultInfo struct {
	TypeName  *types.TypeName
	ErrorType errortypes.ErrorType
}

// Result is the result of the detecttypes analyzer. It contains a list of all
// error types whose pointer-ness could be unambiguously determined.
type Result struct {
	Types []ResultInfo
}

// createResult combines all determined type information into the final analyzer result.
// It merges types from the current package and dependencies (facts) and local overrides,
// with local overrides having the highest precedence.
func (p pass) createResult() Result {
	facts := p.AllObjectFacts()

	// Add types from dependencies (via facts).
	determinedTypes := extractErrorTypes(facts)

	// Iterate over all types in the current package whose pointer-ness has been determined.
	for tn, errorType := range p.AllDetermined {
		if tn.Pkg() == p.Pkg {
			// Export this information as a fact when the type is defined in the current package.
			// These facts can then be consumed by analyzers running on packages dependent on this one.
			p.ExportObjectFact(tn, &errorType)
		}

		// Add type to result.
		// Local overrides will overwrite any existing entries from facts, only for this package.
		determinedTypes[tn] = errorType
	}

	// Convert map to slice for the result.
	return createResult(determinedTypes)
}

func createResult(determinedTypes map[*types.TypeName]errortypes.ErrorType) Result {
	typs := make([]ResultInfo, 0, len(determinedTypes))
	for tn, errorType := range determinedTypes {
		typs = append(typs, ResultInfo{TypeName: tn, ErrorType: errorType})
	}

	return Result{Types: typs}
}

func extractErrorTypes(facts []analysis.ObjectFact) map[*types.TypeName]errortypes.ErrorType {
	determinedTypes := make(map[*types.TypeName]errortypes.ErrorType, len(facts))
	for _, f := range facts {
		fact, ok := f.Fact.(*errortypes.ErrorType)
		if !ok {
			continue
		}

		tn, ok := f.Object.(*types.TypeName)
		if !ok {
			continue
		}

		determinedTypes[tn] = *fact
	}

	return determinedTypes
}
