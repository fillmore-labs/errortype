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
	"fillmore-labs.com/errortype/internal/detectwrappers/result"
	"fillmore-labs.com/errortype/internal/errorlike"
	"fillmore-labs.com/errortype/internal/settings"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// checkAssign creates a new reporter for assignments.
func (p Pass) checkAssign(t types.Type, e ast.Expr) {
	reporter := report.Assign{Base: report.Base{Pass: p.Pass, Expr: e}}
	p.Check(t, reporter)
}

// checkCallArg creates a new reporter for call arguments converted to an error interface.
func (p Pass) checkCallArg(t types.Type, e ast.Expr) {
	reporter := report.CallArg{Base: report.Base{Pass: p.Pass, Expr: e}}
	p.Check(t, reporter)
}

// checkCompositeLit creates a new reporter for composite literal elements converted to an error interface.
func (p Pass) checkCompositeLit(t types.Type, e ast.Expr) {
	reporter := report.CompositeLit{Base: report.Base{Pass: p.Pass, Expr: e}}
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

	if p.Has(settings.OptionStyleCheck) {
		reporter.CheckStyle(t)
	}
}

// checkErrorf creates a new reporter for fmt.Errorf like functions.
func (p Pass) checkErrorf(t types.Type, e ast.Expr) {
	reporter := report.Errorf{Base: report.Base{Pass: p.Pass, Expr: e}}
	p.Check(t, reporter)
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

// checkInstance creates a new reporter for generic instantiations of error-constrained type parameters.
func (p Pass) checkInstance(t types.Type, e ast.Expr) {
	reporter := report.Instance{Base: report.Base{Pass: p.Pass, Expr: e}}
	p.Check(t, reporter)
}

// checkVarDecl creates a new reporter for variable declarations.
// Results of checked type queries are skipped, since a mismatch is reported at the query.
func (p Pass) checkVarDecl(t types.Type, e ast.Expr, name string) {
	if p.isCheckedQuery(e, t) {
		return
	}

	reporter := report.VarDecl{Base: report.Base{Pass: p.Pass, Expr: e}, VarName: name}
	p.Check(t, reporter)
}

// checkSend creates a new reporter for channel send values converted to an error interface.
func (p Pass) checkSend(t types.Type, e ast.Expr) {
	reporter := report.Send{Base: report.Base{Pass: p.Pass, Expr: e}}
	p.Check(t, reporter)
}

// isCheckedQuery reports whether e is a type assertion on an error, a call of an
// errors.AsType-like function, or a generic call returning a checked error-constrained
// type parameter. These are checked themselves, and the type of their result follows
// from the queried or instantiated type.
func (p Pass) isCheckedQuery(e ast.Expr, resultType types.Type) bool {
	switch e := ast.Unparen(e).(type) {
	case *ast.TypeAssertExpr:
		return e.Type != nil && errorlike.IsInterface(p.TypesInfo.Types[e.X].Type)

	case *ast.CallExpr:
		fun, ok := typeutil.FuncOf(p.TypesInfo, e)
		if !ok {
			return false
		}

		if w, ok := p.ErrorFuncs[fun.Func]; ok && (w.Type == result.WrapperAsType || w.Type == result.FuncAssert) {
			return true
		}

		return p.isCheckedTypeParamResult(fun, resultType)

	default:
		return false
	}
}

func (p Pass) isCheckedTypeParamResult(fun typeutil.ResolvedFunc, resultType types.Type) bool {
	inst, ok := p.TypesInfo.Instances[fun.Ident]
	if !ok {
		return false
	}

	sig := typeutil.SignatureOf(fun.Func)
	if sig == nil {
		return false
	}

	tparams := sig.TypeParams()
	for r := range sig.Results().Variables() {
		tp, ok := r.Type().(*types.TypeParam)
		if !ok {
			continue
		}

		if idx := tp.Index(); idx < tparams.Len() && idx < inst.TypeArgs.Len() &&
			tparams.At(idx) == tp &&
			errorlike.IsInterface(tp.Constraint()) &&
			types.Identical(inst.TypeArgs.At(idx), resultType) {
			return true
		}
	}

	return false
}
