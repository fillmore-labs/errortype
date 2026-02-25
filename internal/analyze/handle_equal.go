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

package analyze

import "go/ast"

// handleEqual processes assertion functions that perform comparisons.
//
// It checks if the function is one of the targeted comparison functions
// and delegates the analysis of its arguments to comparison.
func (p Pass) handleEqual(call *ast.CallExpr, argIndex int) {
	if len(call.Args) < 2 { // Other function or multivalued argument
		return
	}

	if argIndex+1 >= len(call.Args) { // should not happen
		p.ReportErrorf(call, "Got only %d arguments, expected at least %d", len(call.Args), argIndex+1)
		return
	}

	// Delegate analysis of errors.Is(..., ...) to comparison.
	p.comparison(call, call.Args[argIndex], call.Args[argIndex+1], true)
}
