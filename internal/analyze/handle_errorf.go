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
	"go/constant"
)

// handleErrorf checks the operands wrapped by %w verbs in a call to a
// fmt.Errorf-style function. srcIdx is the index of the format string in
// call.Args and tgtIdx the index of the first operand.
func (p Pass) handleErrorf(call *ast.CallExpr, srcIdx, tgtIdx int) {
	if call.Ellipsis.IsValid() {
		return // The operands are passed as a slice and can not be matched to verbs.
	}

	val := p.TypesInfo.Types[call.Args[srcIdx]].Value
	if val == nil || val.Kind() != constant.String {
		return // Not a constant format string.
	}

	format := constant.StringVal(val)
	for argNum := range AllWrappedArgs(format, len(call.Args)-tgtIdx) {
		expr := call.Args[tgtIdx+argNum]

		typ, ok := p.TypesInfo.Types[expr]
		if !ok {
			continue
		}

		p.checkErrorf(typ.Type, expr)
	}
}
