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
	"strings"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// handleVarDecls checks for incorrect pointer/value usage of error types in variable declarations.
func (p Pass) handleVarDecls(n *ast.ValueSpec) {
	if len(n.Values) == 0 {
		return // uninitialized variables
	}

	checkType := true

	if n.Type != nil {
		if tv, ok := p.TypesInfo.Types[n.Type]; !ok || !typeutil.HasErrorMethod(tv.Type) {
			return
		}

		checkType = false
	}

	for i, id := range n.Names {
		name := id.Name

		if checkType && p.PrefixFilter() && !hasErrPrefix(id.Name) {
			continue
		}

		typ, value := typeutil.ResultOf(p.TypesInfo, n.Values, i)
		if typ == nil || checkType && !typeutil.HasErrorMethod(typ) {
			continue // Not an error type
		}

		p.checkVarDecl(typ, value, name)
	}
}

// hasErrPrefix checks if a variable name indicates it is intended to be an error variable.
// This heuristic prevents false positives for short-lived local variables that happen to
// hold an error type, ensuring we only enforce semantics on explicit error variables.
func hasErrPrefix(name string) bool {
	return strings.HasPrefix(name, "Err") || strings.HasPrefix(name, "err")
}
