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
	"iter"

	"golang.org/x/tools/go/ast/inspector"
)

// AllReturns iterates over all return statements within the AST tree of the
// node at the given inspector cursor.
//
// It does not descend into nested function literals, so only return statements
// of the current function are considered.
func AllReturns(b inspector.Cursor) iter.Seq[*ast.ReturnStmt] {
	return func(yield func(*ast.ReturnStmt) bool) {
		cont := true

		b.Inspect(
			[]ast.Node{(*ast.FuncLit)(nil), (*ast.ReturnStmt)(nil)},
			func(c inspector.Cursor) bool {
				switch n := c.Node().(type) {
				case *ast.FuncLit:
					return false // Don't check returns in nested function literals

				case *ast.ReturnStmt:
					if cont {
						cont = yield(n)
					}
				}

				return cont
			},
		)
	}
}
