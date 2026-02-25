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
	"go/types"

	"fillmore-labs.com/errortype/internal/astutil"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// handleTypeSpec checks for missing Is or Unwrap methods on not comparable value error types.
func (p Pass) handleTypeSpec(spec *ast.TypeSpec, decl *ast.GenDecl) {
	id := spec.Name
	if id.Name == "_" {
		return
	}

	tn, ok := p.TypesInfo.Defs[id].(*types.TypeName)
	if !ok { // should not happen
		return
	}

	if !p.IsValueError(tn) {
		return
	}

	typ := tn.Type()
	if types.Comparable(typ) {
		return
	}

	if _, ok := astutil.CheckTypeSpecComments(spec, decl, p.Analyzer.Name); !ok {
		return
	}

	// We check for methods defined on either the value or pointer receiver (addressable=true).
	// If a method is defined on a pointer receiver, handleMethodDecl will separately flag it
	// for having the wrong receiver type (et:rcv).
	if typeutil.IsMethod.HasMethod(typ, true) {
		return
	}

	if tn.Exported() {
		p.ReportRangef(spec, "Exported not comparable error type %q should be a pointer type or implement the \"Is\" method. (et:nce+)", tn.Name())
		return
	}

	if typeutil.UnwrapMethod.HasMethod(typ, true) {
		return
	}

	p.ReportRangef(spec, "Unexported not comparable error type %q should be a pointer type or implement the \"Is\" or \"Unwrap\" method. (et:nce+)", tn.Name())
}
