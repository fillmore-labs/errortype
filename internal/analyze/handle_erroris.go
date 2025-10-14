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

import (
	"go/ast"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// handleErrorIs processes function calls like `errors.Is` or assertion functions
// from `github.com/stretchr/testify` that perform error comparisons.
//
// It checks if the function is one of the targeted comparison functions
// and delegates the analysis of its arguments to comparison.
func (p pass) handleErrorIs(n *ast.CallExpr, methodExpr bool, ftyp typeutil.FuncType, checkIs bool) {
	if len(n.Args) < 2 { // Other function or multivalued argument
		return
	}

	baseArg := 0
	if methodExpr {
		baseArg = 1
	}

	switch ftyp {
	case typeutil.IsFunc0:
		if len(n.Args) < 2+baseArg { // should not happen
			p.ReportErrorf(n, "Got only %d arguments, expected at least %d", len(n.Args), 3+baseArg)
			return
		}
		// Delegate analysis of errors.Is(..., ...) to comparison.
		p.comparison(n, n.Args[baseArg], n.Args[baseArg+1], false, checkIs)

	case typeutil.IsFunc1:
		if len(n.Args) < 3+baseArg { // should not happen
			p.ReportErrorf(n, "Got only %d arguments, expected at least %d", len(n.Args), 3+baseArg)
			return
		}

		// Delegate analysis of assert.ErrorIs(t, ..., ...) or assert.Equal(t, ..., ...) to comparison.
		p.comparison(n, n.Args[baseArg+1], n.Args[baseArg+2], false, checkIs)

	default: // should not happen
		p.ReportErrorf(n, "Unconfigured function %d", ftyp)
		return
	}
}
