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

package properties

import (
	"go/types"
	"slices"
	"strings"

	"fillmore-labs.com/errortype/internal/errortypes"
)

// Properties is a map that associates a type name with its corresponding ErrorProperty, describing error type behaviors.
type Properties map[*types.TypeName]ErrorProperty

// HasUndeterminedErrors checks if the PropertyMap contains any entries with undetermined error types.
func (p Properties) HasUndeterminedErrors() bool {
	for _, errorType := range p {
		if errorType.DeterminedType() == errortypes.UndecidedType {
			return true
		}
	}

	return false
}

// AllSorted is an iterator over the DetectedTypes map in sorted order of type names.
// Sorting is based on package paths and names of the types.
func (p Properties) AllSorted(yield func(*types.TypeName, ErrorProperty) bool) {
	typeNames := make([]*types.TypeName, 0, len(p))
	for tn := range p {
		typeNames = append(typeNames, tn)
	}

	slices.SortFunc(typeNames, compareTypeName)

	for _, tn := range typeNames {
		if !yield(tn, p[tn]) {
			return
		}
	}
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
