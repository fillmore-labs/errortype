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

// processWrappers finds functions wrapping error type checking functions like [errors.As] and [errors.Is].
func (p pass) processWrappers(ctx context.Context, overrides WrapperOverrides) {
	defer trace.StartRegion(ctx, "wrappers").End()

	pkg := p.Pkg

	// Seed with known functions.
	if seeds := wrappers.LookupSeeds(pkg); len(seeds) > 0 {
		p.exportWrappers(seeds)
	}

	var pkgOverrides []wrappers.ExplicitWrapper

	for name, wrapperType := range overrides {
		if name.Path != pkg.Path() {
			continue
		}

		pkgOverrides = append(pkgOverrides, wrappers.ExplicitWrapper{LocalFuncName: name.LocalFuncName, Typ: wrapperType})
	}

	var funcs result.ErrorFuncs
	if len(pkgOverrides) > 0 {
		// Skip autodetection when the package is overridden
		funcs = wrappers.LookupWrappers(pkg.Scope(), pkgOverrides)
	} else {
		funcs = wrappers.DetectWrappers(p.Pass, p.ErrorFuncs)
	}

	p.exportWrappers(funcs)
}

// exportWrappers exposes the given wrapped functions to downstream analyzers.
func (p pass) exportWrappers(funcs result.ErrorFuncs) {
	for fun, wrapper := range funcs {
		p.ErrorFuncs[fun] = wrapper
		p.ExportObjectFact(fun, &wrapper)
	}
}
