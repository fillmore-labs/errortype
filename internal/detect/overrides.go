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
	"strings"

	"fillmore-labs.com/errortype/internal/errortypes"
	"fillmore-labs.com/errortype/internal/overrides"
	"fillmore-labs.com/errortype/internal/typeutil"
)

func (o *options) addOverrides(overrides []overrides.Override) {
	if o.usageOverrides == nil {
		o.usageOverrides = make(map[string]map[string]errortypes.ErrorType)
	}

	for _, override := range overrides {
		names, ok := o.usageOverrides[override.Path]
		if !ok {
			names = make(map[string]errortypes.ErrorType)
			o.usageOverrides[override.Path] = names
		}

		names[override.Name] = override.ErrorType
	}
}

func (p pass) processOverrides(overrides map[string]map[string]errortypes.ErrorType) {
	pkg := p.Pkg

	path, scope := pkg.Path(), pkg.Scope()
	for name, usage := range overrides[path] {
		// Look up and ensure the object is found and actually a type name.
		tn, ok := scope.Lookup(name).(*types.TypeName)
		if !ok {
			if !p.hasTestFiles() { // could be defined in test files
				log.Printf("Can't find override %q in package %s", name, path)
			}

			continue
		}

		var property ErrorProperty

		// Check whether the override is valid.
		switch usage {
		case errortypes.PointerType:
			ptrType := types.NewPointer(tn.Type())
			if !typeutil.HasErrorMethod(ptrType) {
				log.Printf("Pointer override \"*%s\" does not implement the error interface", name)

				continue
			}
			property = PointerOverride

		case errortypes.ValueType:
			if !typeutil.HasErrorMethod(tn.Type()) {
				log.Printf("Value override \"%s\" does not implement the error interface", name)

				continue
			}
			property = ValueOverride

		case errortypes.SuppressType:
			property = SuppressOverride

		default: // should not happen
			continue
		}

		old := p.AddTypeProperty(tn, property)

		if old&OverrideMask == None {
			if prop := old.DeterminedType(); prop == usage {
				log.Printf("Redundant override: %s %d", name, old)
			}
		}
	}
}

// hasTestFiles checks if any file in the pass has a _test.go suffix.
func (p pass) hasTestFiles() bool {
	for file := range p.Fset.Iterate {
		if strings.HasSuffix(file.Name(), "_test.go") {
			return true
		}
	}

	return false
}
