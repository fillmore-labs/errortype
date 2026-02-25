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

package result

import "go/types"

type (
	// ErrorTypes is a map associating *[types.TypeName] with its detected [ErrorType].
	ErrorTypes map[*types.TypeName]ErrorType

	// ErrorFuncs maps functions to their corresponding error wrapper function metadata.
	ErrorFuncs map[types.Object]ErrorFunc

	// Result is the result of the detecttypes analyzer. It contains a list of all
	// error types whose pointer-ness could be unambiguously determined.
	Result struct {
		// Types holds the determined pointer-ness for a type, identified by its *[types.TypeName].
		Types ErrorTypes
		// Funcs describes detected wrapper functions.
		Funcs ErrorFuncs
	}
)

// New creates a [Result] by converting a map.
func New(types ErrorTypes, funcs ErrorFuncs) Result {
	return Result{Types: types, Funcs: funcs}
}
