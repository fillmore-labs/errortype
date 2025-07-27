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

import "golang.org/x/tools/go/analysis"

// run is the main function for the detecttypes analyzer.
//
// It inspects type, function and variable declarations to infer whether an error type
// is intended to be used as a pointer or a value, including handling
// local and usage overrides.
//
// It then exports the determined properties as facts for downstream packages and
// returns a result containing all relevant properties for the current analysis pass.
func (o *options) run(ap *analysis.Pass) (any, error) {
	p := newPass(ap)

	// Process type declarations in the current package.
	p.processTypeDecls()

	// Process variable declarations, identifying properties for both local and external types.
	// External type properties are considered local overrides.
	p.processVarSpecs()

	// Calculate overrides and log impossible ones.
	p.processOverrides(o.usageOverrides)

	if o.heuristics&HeuristicUsage != 0 && p.HasUndeterminedErrors() {
		// Process error value usage in the current package.
		p.processUsage()
	}

	if o.heuristics&HeuristicReceivers != 0 && p.HasUndeterminedErrors() {
		// Last resort.
		p.processReceivers()
	}

	// Process alias declarations in the current package.
	p.processAliases()

	if o.debug {
		p.logResults()
	}

	// Export determined properties for types in the current package as facts for downstream packages.
	// Create and return a result containing all determined properties for the current analysis pass,
	// including those from dependencies (facts), the current package, and local overrides.
	result := p.createResult()

	return result, nil
}
