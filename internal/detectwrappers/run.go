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

package detectwrappers

import (
	"context"
	"log"
	"runtime/trace"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/settings/errorwrapper"
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
func Run(pass *analysis.Pass, o *errorwrapper.Config) (any, error) {
	ctx := context.Background()

	ctx, task := trace.NewTask(ctx, "detecttypes")
	defer task.End()

	if trace.IsEnabled() {
		trace.Log(ctx, "pkg", typeutil.PkgPath(pass.Fset, pass.Pkg))
	}

	// Add types from dependencies (via facts).
	errorFuncs := extractFacts(pass)

	p := newPass(pass, errorFuncs)

	debug := o.Trace != nil && o.Trace.MatchString(pass.Pkg.Path())

	// Detect error wrappers in the current package.
	p.processWrappers(ctx, o.WrapperOverrides, debug)

	if debug {
		p.logResults(ctx, log.Default())
	}

	// Export determined properties for types in the current package as facts for downstream packages.
	// Create and return a result containing all determined properties for the current analysis pass,
	// including those from dependencies (facts), the current package, and local overrides.
	return p.createResult(ctx), nil
}
