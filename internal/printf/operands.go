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

// Package printf parses printf-style format strings to find the operands
// wrapped by %w verbs in calls to fmt.Errorf-like functions.
package printf

import (
	"go/ast"
	"go/constant"
	"go/types"
	"iter"
)

// WrappedOperands returns the operands wrapped by %w verbs, given the format string
// expression and the operands of a fmt.Errorf-like call. It yields nothing when the
// format string is not a constant.
func WrappedOperands(info *types.Info, formatExpr ast.Expr, operands []ast.Expr) iter.Seq[ast.Expr] {
	return func(yield func(ast.Expr) bool) {
		val := info.Types[formatExpr].Value
		if val == nil || val.Kind() != constant.String {
			return // Not a constant format string.
		}

		for argNum := range AllWrappedArgs(constant.StringVal(val), len(operands)) {
			if !yield(operands[argNum]) {
				return
			}
		}
	}
}
