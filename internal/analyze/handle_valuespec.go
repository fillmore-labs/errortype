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
	"go/types"
	"strings"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// handleValueSpec checks for incorrect pointer/value usage of error types in variable declarations.
func (p Pass) handleValueSpec(spec *ast.ValueSpec, packageLevel bool) {
	if len(spec.Values) == 0 {
		return // uninitialized variables
	}

	checkType := true

	if spec.Type != nil {
		if tv, ok := p.TypesInfo.Types[spec.Type]; !ok || !typeutil.HasErrorMethod(tv.Type) {
			return
		}

		checkType = false
	}

	for i, id := range spec.Names {
		name := id.Name

		typ, value := typeutil.ResultOf(p.TypesInfo, spec.Values, i)
		if typ == nil || checkType && !typeutil.HasErrorMethod(typ) {
			continue // Not an error type
		}

		if packageLevel && name != "_" && !types.Comparable(typ) && !typeutil.IsMethod.HasMethod(typ, true) {
			typeName := types.TypeString(typ, types.RelativeTo(p.Pkg))
			p.ReportRangef(id, "Not comparable error sentinel %q of type %q should implement the \"Is\" method. (et:nce)", name, typeName)
		}

		if checkType && p.Has(OptionPrefixFilter) && !conformantVar(id.Name) {
			continue
		}

		p.checkVarDecl(typ, value, name)
	}
}

const (
	// prefix for unexported error variables.
	varPrefix = "err"

	// Suffix for unexported error variables.
	varSuffix = "Err"

	// prefix for exported error variables.
	varPrefixExported = "Err"
)

// conformantVar checks if a variable name indicates it is intended to be an error variable.
// This heuristic prevents false positives for short-lived local variables that happen to
// hold an error type, ensuring we only enforce semantics on explicit error variables.
func conformantVar(name string) bool {
	if token.IsExported(name) {
		return strings.HasPrefix(name, varPrefixExported)
	}

	return strings.HasPrefix(name, varPrefix) || strings.HasSuffix(name, varSuffix)
}
