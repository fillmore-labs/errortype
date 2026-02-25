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
	"runtime/trace"

	"fillmore-labs.com/errortype/detect/result"
	"fillmore-labs.com/errortype/internal/detect/properties"
)

// processAliases transfers properties from aliased types to the aliases themselves.
// Aliases do not have their own methods or properties, but inherit the behavior of the type they alias.
func (p pass) processAliases(ctx context.Context, eTypes result.ErrorTypes) {
	defer trace.StartRegion(ctx, "aliases").End()

	for alias := range p.ErrorTypes {
		if !alias.IsAlias() {
			continue // We are only interested in aliases.
		}

		// Resolve the alias to its underlying named type.
		var tn *types.TypeName

		switch n := types.Unalias(alias.Type()).(type) {
		case *types.Named:
			tn = n.Obj()

		case *types.Pointer:
			continue // Already done in type declaration

		default:
			continue // Alias to an unnamed type with embedded error
		}

		var property properties.ErrorProperty
		if oldp, ok := p.ErrorTypes[tn]; ok {
			// If the original type is in the same package, its properties
			// have already been computed by processTypeDecls. We can copy them.
			property = oldp &^ properties.OverrideMask // Copy all but override flags.
		} else {
			// If the original type is from another package, we rely on
			// the facts exported by that package's analysis.
			errorType, ok := eTypes[tn]
			if !ok {
				continue
			}

			switch errorType & result.ExpectedMask {
			case result.Pointer:
				property = properties.PointerAlias

			case result.Value:
				property = properties.ValueAlias

			default: // Undecided or suppressed
				continue
			}
		}

		// We have either found the type in our property map or imported an ErrorType fact.
		p.ErrorTypes[alias] |= property
	}
}
