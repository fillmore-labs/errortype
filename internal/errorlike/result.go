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

package errorlike

import (
	"go/ast"
	"go/types"
)

// ResultIndex checks whether the given function result list has an error type as its last return value.
// Returns the index of the error result or -1 when not found.
func ResultIndex(info *types.Info, ftype *ast.FuncType) int {
	results := ftype.Results

	// We are only interested in functions with return values.
	if results == nil || len(results.List) == 0 {
		return -1 // No result
	}

	// Only check the last return type expression, as `error` is
	// conventionally the last one.
	lastType := results.List[len(results.List)-1].Type

	// Check if the return type is a type with an `Error() string` method.
	if tv, ok := info.Types[lastType]; ok && IsInterface(tv.Type) {
		return results.NumFields() - 1
	}

	return -1 // Not an error type
}
