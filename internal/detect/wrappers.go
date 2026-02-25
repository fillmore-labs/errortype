// Copyright 2026 Oliver Eikemeier. All Rights Reserved.
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
	"runtime/trace"

	"fillmore-labs.com/errortype/detect/result"
	"fillmore-labs.com/errortype/internal/detect/wrappers"
)

// WrapperOverrides maps a package path to a map of function names to their corresponding error wrapper metadata.
type WrapperOverrides map[string][]wrappers.ExplicitWrapper

// processWrappers finds functions wrapping error type checking functions like [errors.As] and [errors.Is].
func (d dpass) processWrappers(ctx context.Context, overrides WrapperOverrides, debug bool) {
	defer trace.StartRegion(ctx, "wrappers").End()

	pkg := d.Pkg

	// Seed with known functions for this package (e.g. the standard "errors" package).
	if seeds := wrappers.LookupSeeds(pkg); len(seeds) > 0 {
		d.exportWrappers(seeds)
	}

	// Resolve explicit overrides for this package. These are authoritative: they
	// participate as known wrappers during autodetection but are not themselves
	// re-derived as candidates.
	if pkgOverrides := overrides[pkg.Path()]; len(pkgOverrides) > 0 {
		d.exportWrappers(wrappers.LookupWrappers(pkg, pkgOverrides, debug))
	}

	// Autodetect wrappers, treating imported facts, seeds, and overrides (all in p.ErrorFuncs)
	// as the known set. Functions already classified are skipped as candidates.
	d.exportWrappers(wrappers.DetectWrappers(d.TypesInfo, d.Files, d.ErrorFuncs))
}

// exportWrappers exposes the given wrapped functions to downstream analyzers.
func (d dpass) exportWrappers(funcs result.ErrorFuncs) {
	for fun, wrapper := range funcs {
		d.ErrorFuncs[fun] = wrapper
		d.ExportObjectFact(fun, &wrapper)
	}
}
