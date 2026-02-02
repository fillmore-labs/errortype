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

package usage

import (
	"go/types"

	"fillmore-labs.com/errortype/facts"
)

// ErrorUsage maps error types to their usage information.
// It is used to analyze and track the observed and expected usage of error types.
type ErrorUsage map[*types.TypeName]Usage

// allDetermined is an iterator over all types in the map whose pointer-ness
// has been unambiguously determined (i.e., where DeterminedType returns true).
// The iterator yields the type's TypeName and a boolean indicating if it's a pointer type.
func (e ErrorUsage) allDetermined(yield func(*types.TypeName, facts.ErrorFact) bool) {
	for tn, usage := range e {
		if typ := usage.DeterminedType(); typ != facts.UndecidedType {
			if !yield(tn, typ) {
				return
			}
		}
	}
}
