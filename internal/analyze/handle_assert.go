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

	"fillmore-labs.com/errortype/internal/typeutil"
)

// handleTypeAssert checks for incorrect pointer/value usage of error types in type assertions.
func (p Pass) handleTypeAssert(n *ast.TypeAssertExpr) {
	if n.Type == nil {
		return // Type switches are handled in handleTypeSwitch
	}

	eTyp := p.TypesInfo.Types[n.X].Type
	if !typeutil.IsInterfaceWithError(eTyp) {
		return // We are only interested in assertions on interfaces that implement error.
	}

	// The type is of course n.Type, but we can check whether it is used in the special form "v, ok"
	typ := p.TypesInfo.Types[n].Type

	if t, ok := typ.(*types.Tuple); ok {
		typ = t.At(0).Type()
	} else if p.UncheckedAssert() {
		p.reportUnchecked(n, eTyp, typ)
	}

	p.checkAssert(typ, n.Type)
}

func (p Pass) reportUnchecked(n *ast.TypeAssertExpr, eTyp, typ types.Type) {
	ityp, isInterface := typ.Underlying().(*types.Interface)
	if isInterface && types.Implements(eTyp, ityp) {
		return
	}

	// Distinguish between assertions to a concrete type vs. an interface type for the error code.
	codeSuffix := ""
	if isInterface {
		codeSuffix = "+"
	}

	name := types.TypeString(typ, types.RelativeTo(p.Pkg))

	p.ReportRangef(n, "Asserting error to %q without checking might lead to a run-time panic. (et:uca%s)", name, codeSuffix)
}
