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
	"log"

	"fillmore-labs.com/errortype/detect/result"
	"fillmore-labs.com/errortype/internal/detect/properties"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// UsageOverrides maps a fully qualified type name to its corresponding error usage type.
type UsageOverrides map[typeutil.TypeName]result.ErrorType

// processOverrides applies error usage overrides to the property map, validating and logging invalid configurations.
func (d dpass) processOverrides(overrides UsageOverrides) {
	for tn, property := range d.ErrorTypes {
		typeName := typeutil.NewTypeName(tn)

		usage, ok := overrides[typeName]
		if !ok {
			continue
		}

		// Check whether the override is valid.
		switch usage {
		case result.Pointer:
			if ptrType := types.NewPointer(tn.Type()); !typeutil.HasErrorMethod(ptrType) {
				log.Printf("Pointer override \"*%s\" does not implement the error interface", typeName)

				continue
			}
			property |= properties.PointerOverride

		case result.Value:
			if !typeutil.HasErrorMethod(tn.Type()) {
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

		d.ErrorTypes[tn] = property
	}
}
