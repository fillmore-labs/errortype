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

package detecttypes

import (
	"go/types"
	"log"

	"fillmore-labs.com/errortype/internal/detecttypes/properties"
	"fillmore-labs.com/errortype/internal/detecttypes/result"
	"fillmore-labs.com/errortype/internal/errorlike"
	"fillmore-labs.com/errortype/internal/qualname"
	"fillmore-labs.com/errortype/internal/settings/errortype"
)

// processOverrides applies error usage overrides to the property maps, validating and logging invalid configurations.
func (d dpass) processOverrides(overrides errortype.UsageOverrides, local *localAssertions) {
	applyOverrides(d.ErrorTypes, overrides)

	if local != nil && len(local.types) > 0 {
		applyOverrides(local.types, overrides)
	}
}

func applyOverrides(props properties.Properties, overrides errortype.UsageOverrides) {
	for tn, property := range props {
		typeName := qualname.NewType(tn)

		usage, ok := overrides[typeName]
		if !ok {
			continue
		}

		property &^= properties.OverrideMask

		// Check whether the override is valid.
		switch usage {
		case result.Pointer:
			if ptrType := types.NewPointer(tn.Type()); !errorlike.Implements(ptrType) {
				log.Printf("Pointer override \"*%s\" does not implement the error interface", typeName)

				continue
			}

			property |= properties.PointerOverride

		case result.Value:
			if !errorlike.Implements(tn.Type()) {
				log.Printf("Value override %q does not implement the error interface", typeName)

				continue
			}

			property |= properties.ValueOverride

		case result.Suppress:
			property |= properties.SuppressOverride

		default: // should not happen
			log.Printf("Unknown override type %s for %q", usage, typeName)

			continue
		}

		props[tn] = property
	}
}
