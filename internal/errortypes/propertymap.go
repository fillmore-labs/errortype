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

package errortypes

import (
	"go/types"
	"slices"
	"strings"

	"golang.org/x/exp/constraints"
)

// TypeProperty in an interface for integer-based property flags with a method to determine
// the resolved error type.
type TypeProperty interface {
	constraints.Integer
	DeterminedType() ErrorType
}

// PropertyMap stores the collected properties for each type encountered during analysis.
// It is keyed by the [*types.TypeName] of the type's definition to ensure uniqueness.
type PropertyMap[P TypeProperty] map[*types.TypeName]P

// NewPropertyMap creates a new, empty PropertyMap.
func NewPropertyMap[P TypeProperty]() PropertyMap[P] {
	return make(PropertyMap[P])
}

// GetTypeProperty retrieves the property associated with the given TypeName in the map.
// It returns the property and a boolean indicating if the TypeName exists in the map.
func (p PropertyMap[P]) GetTypeProperty(tn *types.TypeName) (P, bool) {
	old, ok := p[tn]
	return old, ok
}

// SetTypeProperty assigns a given property to a TypeName in the PropertyMap, overwriting any existing property.
func (p PropertyMap[P]) SetTypeProperty(tn *types.TypeName, property P) {
	p[tn] = property
}

// AddTypeProperty adds a property to a type in the map.
// It combines the new property with any existing properties for the given type.
// If the type is not yet in the map, it is added.
func (p PropertyMap[P]) AddTypeProperty(tn *types.TypeName, newProperty P) P {
	old, ok := p.GetTypeProperty(tn)

	var unset P
	if !ok || old&newProperty == unset { // properties are not set.
		p.SetTypeProperty(tn, old|newProperty)
	}

	return old
}

// HasUndeterminedErrors checks if the PropertyMap contains any entries with undetermined error types.
func (p PropertyMap[P]) HasUndeterminedErrors() bool {
	for _, errorType := range p {
		if errorType.DeterminedType() == Undecided {
			return true
		}
	}

	return false
}

// AllDetermined is an iterator over all types in the map whose pointer-ness
// has been unambiguously determined (i.e., where DeterminedType returns true).
// The iterator yields the type's TypeName and a boolean indicating if it's a pointer type.
func (p PropertyMap[P]) AllDetermined(yield func(*types.TypeName, ErrorType) bool) {
	for tn, info := range p {
		if typ := info.DeterminedType(); typ != Undecided {
			if !yield(tn, typ) {
				return
			}
		}
	}
}

// AllSorted is an iterator over all entries in the PropertyMap in sorted order by type name and package path.
func (p PropertyMap[P]) AllSorted(yield func(*types.TypeName, P) bool) {
	typeNames := p.sortedTypeNames()

	for _, tn := range typeNames {
		info := p[tn]
		if !yield(tn, info) {
			return
		}
	}
}

// sortedTypeNames returns a sorted slice of type names from the given PropertyMap.
// Sorting is based on package path and type name.
func (p PropertyMap[P]) sortedTypeNames() []*types.TypeName {
	typeNames := make([]*types.TypeName, 0, len(p))
	for tn := range p {
		typeNames = append(typeNames, tn)
	}

	slices.SortFunc(typeNames, compareTypeName)

	return typeNames
}

// compareTypeName compares two type names based on their package paths and names.
func compareTypeName(a, b *types.TypeName) int {
	var patha, pathb string
	if pkg := a.Pkg(); pkg != nil {
		patha = pkg.Path()
	}

	if pkg := b.Pkg(); pkg != nil {
		pathb = pkg.Path()
	}

	if i := strings.Compare(patha, pathb); i != 0 {
		return i
	}

	return strings.Compare(a.Name(), b.Name())
}
