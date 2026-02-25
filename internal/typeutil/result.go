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

package typeutil

import (
	"go/ast"
	"go/types"
)

// ResultOf extracts the expression and its inferred type for a return value at a given index.
// It consistently handles single expressions, multiple return expressions, and tuple returns.
func ResultOf(info *types.Info, exprs []ast.Expr, idx int) (types.Type, ast.Expr) {
	switch l := len(exprs); l {
	case 0:
		return nil, nil

	case 1:
		expr := exprs[0]

		typ := info.Types[expr].Type
		if tuple, ok := typ.(*types.Tuple); ok {
			if idx >= tuple.Len() {
				return nil, expr // Out of bounds tuple
			}

			return tuple.At(idx).Type(), expr
		}

		if idx != 0 {
			return nil, expr // Single value addressed at > 0
		}

		return typ, expr // Single value at index 0

	default:
		if idx >= l {
			return nil, nil // Multiple values with index > len
		}

		expr := exprs[idx]

		return info.Types[expr].Type, expr
	}
}
