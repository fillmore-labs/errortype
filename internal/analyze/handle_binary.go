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
	"go/token"
)

// handleBinaryExpr checks binary expressions for equality or inequality
// comparisons involving addresses of composite literals or new() calls.
func (p pass) handleBinaryExpr(n *ast.BinaryExpr) {
	switch n.Op {
	case token.EQL, token.NEQ:
		p.comparison(n, n.X, n.Y, false)

	default:
		return
	}
}
