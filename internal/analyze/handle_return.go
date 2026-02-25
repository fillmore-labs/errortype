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
	"go/types"

	"golang.org/x/tools/go/ast/inspector"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// handleReturns identifies function return parameters that are of type error
// and then inspects all return statements within the function body to check
// for incorrect error type usage.
func (p Pass) handleReturns(body inspector.Cursor, errResultIdx int) {
	body.Inspect([]ast.Node{(*ast.FuncLit)(nil), (*ast.ReturnStmt)(nil)}, func(c inspector.Cursor) (descend bool) {
		switch n := c.Node().(type) {
		case *ast.FuncLit:
			return false // Don't check returns in nested function literals, they will be handled separately.

		case *ast.ReturnStmt:
			typ, res := typeutil.ResultOf(p.TypesInfo, n.Results, errResultIdx)
			if typ == nil || typ == types.Typ[types.UntypedNil] {
				return false // Bare return or nil
			}

			p.checkReturn(typ, res)

			return false

		default: // should not happen
			p.ReportErrorf(n, "Unexpected node type: %T", n)

			return true
		}
	})
}
