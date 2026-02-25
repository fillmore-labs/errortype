// Copyright 2026 Oliver Eikemeier. All Rights Reserved.
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

	"fillmore-labs.com/errortype/internal/printf"
)

// handleErrorf checks the operands wrapped by %w verbs in a call to a
// fmt.Errorf-style function. formatExpr is the format string expression and
// operands are the argument expressions starting at the first format operand.
func (p Pass) handleErrorf(formatExpr ast.Expr, operands []ast.Expr) {
	for expr := range printf.WrappedOperands(p.TypesInfo, formatExpr, operands) {
		if typ, ok := p.TypesInfo.Types[expr]; ok {
			p.checkErrorf(typ.Type, expr)
		}
	}
}
