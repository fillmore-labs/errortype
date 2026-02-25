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

	"fillmore-labs.com/errortype/detect/result"
)

// ErrorUsage maps error types to their usage information.
// It is used to analyze and track the observed and expected usage of error types.
type ErrorUsage map[*types.TypeName]Usage

// ProcessDetectedTypes populates the initial error usage map based on the results
// from the prerequisite `detecttypes` analyzer.
func (e ErrorUsage) ProcessDetectedTypes(eTypes result.ErrorTypes) {
	for tn, errorType := range eTypes {
		var usage Usage

		switch errorType & result.ExpectedMask {
		case result.Pointer:
			usage = PointerExpected

		case result.Value:
			usage = ValueExpected

		case result.Suppress:
			usage = SuppressExpected

		default:
			continue
		}

		e[tn] = usage
	}
}

// IsValueError returns whether the given type name should be a value error.
func (e ErrorUsage) IsValueError(tn *types.TypeName) bool {
	usage, ok := e[tn]

	return ok && usage&ExpectedMask == ValueExpected
}
