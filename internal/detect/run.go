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
	"log"
	"runtime/trace"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/detect/result"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// Run is the main function for the detecttypes analyzer.
//
// It inspects type, function, and variable declarations to infer whether an error type
// is intended to be used as a pointer or a value, including handling
// local and usage overrides.
//
// It then exports the determined properties as facts for downstream packages and
// returns a result containing all relevant properties for the current analysis pass.
func (o *Options) Run(ap *analysis.Pass) (any, error) {
	ctx := context.Background()

	ctx, task := trace.NewTask(ctx, "detecttypes")
	defer task.End()

	if trace.IsEnabled() {
		trace.Log(ctx, "pkg", typeutil.PkgPath(ap))
	}

	p := newPass(ap)

	// Add types from dependencies (via facts).
	var eTypes result.ErrorTypes
	eTypes, p.ErrorFuncs = p.extractFacts()

	// Detect error wrappers in the current package.
	p.processWrappers(ctx, o.WrapperOverrides)

	// Process type declarations in the current package.
	p.processTypeDecls(ctx)

	// Calculate overrides and log impossible ones.
	if len(o.UsageOverrides) > 0 {
		p.processOverrides(o.UsageOverrides)
	}

	if o.Heuristics&HeuristicVar != 0 && p.ErrorTypes.HasUndeterminedErrors() {
		// Process variable declarations, identifying properties for local types.
		p.processVarDecls(ctx)
	}

	if o.Heuristics&HeuristicUsage != 0 && p.ErrorTypes.HasUndeterminedErrors() {
		// Process error value usage in the current package.
		p.processUsage(ctx)
	}

	if o.Heuristics&HeuristicReceivers != 0 && p.ErrorTypes.HasUndeterminedErrors() {
		// Last resort.
		p.processReceivers(ctx)
	}

	// Process alias declarations in the current package.
	p.processAliases(ctx, eTypes)

	if o.Trace != nil && o.Trace.MatchString(p.Pkg.Path()) {
		p.logResults(ctx, log.Default())
	}

	// Export determined properties for types in the current package as facts for downstream packages.
	// Create and return a result containing all determined properties for the current analysis pass,
	// including those from dependencies (facts), the current package, and local overrides.
	return p.createResult(ctx, eTypes), nil
}
