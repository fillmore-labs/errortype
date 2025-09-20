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
func (p pass) handleTypeAssert(n *ast.TypeAssertExpr, uncheckedAssert bool) {
	if n.Type == nil {
		return // Type switches are handled in handleTypeSwitch
	}

	tvx, ok := p.TypesInfo.Types[n.X]
	if !ok || !typeutil.HasErrorMethod(tvx.Type) {
		return // We are only interested in assertions on interface that implement error.
	}

	tv, ok := p.TypesInfo.Types[n]
	if !ok { // should not happen
		return
	}

	var typ types.Type
	if t, ok := tv.Type.(*types.Tuple); ok {
		if t.Len() != 2 || t.At(1).Type() != basicBool { // should not happen
			p.ReportErrorf(n, "Unrecognized tuple structure: %v", t)
			return
		}

		typ = t.At(0).Type()
	} else {
		if uncheckedAssert {
			p.reportUnchecked(n, tvx, tv)
		}

		typ = tv.Type
	}

	p.checkErrorUsage(typ, p.AssertReporter(n.Type))
}

var basicBool = types.Typ[types.Bool]

func (p pass) reportUnchecked(n *ast.TypeAssertExpr, tvx, tv types.TypeAndValue) {
	ityp, isInterface := tv.Type.Underlying().(*types.Interface)
	if isInterface && types.Implements(tvx.Type, ityp) {
		return
	}

	// Distinguish between assertions to a concrete type vs. an interface type for the error code.
	codeSuffix := ""
	if isInterface {
		codeSuffix = "+"
	}

	name := types.TypeString(tv.Type, types.RelativeTo(p.Pkg))

	p.ReportRangef(n, "Asserting error to %q without checking might lead to a run-time panic. (et:uca%s)", name, codeSuffix)
}
