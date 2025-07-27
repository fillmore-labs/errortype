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

	"fillmore-labs.com/errortype/internal/errortypes"
)

// processAliases transfers properties from aliased types to the aliases themselves.
// Aliases do not have their own methods or properties, but inherit the behavior of the type they alias.
func (p pass) processAliases() {
	for alias := range p.PropertyMap {
		if !alias.IsAlias() {
			continue // We are only interested in aliases.
		}

		// Resolve the alias to its underlying named type.
		named, ok := types.Unalias(alias.Type()).(*types.Named)
		if !ok {
			continue // Alias to an unnamed type with embedded error
		}

		orig := named.Obj()

		var property ErrorProperty
		// Check if the original type is from the same package.
		if orig.Pkg() == p.Pkg {
			// If the original type is in the same package, its properties
			// have already been computed by processTypeDecls. We can copy them.
			p, ok := p.GetTypeProperty(orig)
			if !ok {
				continue
			}

			property = p &^ OverrideMask // Copy all but override flags.
		} else {
			// If the original type is from another package, we rely on
			// the facts exported by that package's analysis.
			var errorType errortypes.ErrorType
			if !p.ImportObjectFact(orig, &errorType) {
				continue
			}

			switch errorType {
			case errortypes.PointerType:
				property = PointerAlias

			case errortypes.ValueType:
				property = ValueAlias
			}
		}

		// We have either found the type in our property map, or imported an ErrorType fact
		// so it has to be an alias to an error type.
		p.AddTypeProperty(alias, property)
	}
}
