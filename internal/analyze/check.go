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

	"fillmore-labs.com/errortype/internal/analyze/report"
)

// checkAssign creates a new reporter for assignments.
func (p Pass) checkAssign(t types.Type, e ast.Expr) {
	reporter := report.Assign{Base: report.Base{Pass: p.Pass, Expr: e}}
	p.Check(t, reporter)
}

// checkAssert creates a new reporter for assertions.
func (p Pass) checkAssert(t types.Type, e ast.Expr) {
	reporter := report.Assert{Base: report.Base{Pass: p.Pass, Expr: e}}
	p.Check(t, reporter)
}

// checkErrorsAs creates a new reporter for errors.As like functions.
func (p Pass) checkErrorsAs(t types.Type, e, fun ast.Expr) {
	reporter := report.ErrorsAs{Base: report.Base{Pass: p.Pass, Expr: e}, Fun: fun}
	p.Check(t, reporter)

	if p.StyleCheck() {
		reporter.CheckStyle(t)
	}
}

// checkReturn creates a new reporter for return statements.
func (p Pass) checkReturn(t types.Type, e ast.Expr) {
	reporter := report.Return{Base: report.Base{Pass: p.Pass, Expr: e}}
	p.Check(t, reporter)
}

// checkTypeSwitch creates a new reporter for type switches.
func (p Pass) checkTypeSwitch(t types.Type, e ast.Expr) {
	reporter := report.TypeSwitch{Base: report.Base{Pass: p.Pass, Expr: e}}
	p.Check(t, reporter)
}

// checkGenericCall creates a new reporter for generic functions.
func (p Pass) checkGenericCall(t types.Type, e, fun ast.Expr) {
	reporter := report.GenericCall{Base: report.Base{Pass: p.Pass, Expr: e}, Fun: fun}
	p.Check(t, reporter)
}

// checkVarDecl creates a new reporter for variable declarations.
func (p Pass) checkVarDecl(t types.Type, e ast.Expr, name string) {
	reporter := report.VarDecl{Base: report.Base{Pass: p.Pass, Expr: e}, VarName: name}
	p.Check(t, reporter)
}
