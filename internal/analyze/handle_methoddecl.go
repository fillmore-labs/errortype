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
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/edge"
	"golang.org/x/tools/go/ast/inspector"

	"fillmore-labs.com/errortype/internal/analyze/usage"
	"fillmore-labs.com/errortype/internal/astutil"
	"fillmore-labs.com/errortype/internal/typeutil"
)

func (p Pass) handleMethodDecl(c inspector.Cursor, f *ast.FuncDecl) {
	matchSignature := typeutil.SignatureCheckFor(f.Name.Name)
	if matchSignature == nil { // Not an error-related function
		return
	}

	fun, ok := p.TypesInfo.Defs[f.Name].(*types.Func)
	if !ok { // should not happen
		return
	}

	sig := fun.Signature()
	if sig.Recv() == nil { // should not happen, checked before
		return
	}

	tn, ptr, ok := typeutil.TypeNameOf(types.Unalias(sig.Recv().Type()))
	if !ok { // should not happen
		return
	}

	prop, ok := p.ErrorUsage[tn]
	if !ok {
		return // not an error type
	}

	if astutil.HasNoLint(f.Doc, p.Analyzer.Name) {
		return
	}

	if !matchSignature(sig) {
		// also found by golang.org/x/tools/go/analysis/passes/stdmethods.
		p.ReportRangef(f.Type, "Method %q has the wrong signature (et:sig)", fun.Name())

		return
	}

	if ptr && prop&usage.ExpectedMask == usage.ValueExpected {
		p.ReportRangef(f.Recv, "Method %q should be implemented with a value receiver, not a pointer (et:rcv)", fun.Name())
	}

	// check for an "Is" method declaration and hand over to body analysis.
	if f.Name.Name == typeutil.IsName && f.Body != nil {
		// TODO: unify this with the checkIsMethod below
		param := singleParam(p.TypesInfo, f.Type.Params)
		if param == nil {
			return // An "Is" method without named target parameter is legal, but rare.
		}

		body := c.ChildAt(edge.FuncDecl_Body, -1)
		p.handleIs(body, param)

		if !ptr || !typeutil.ZeroSized(tn.Type()) { // make an exception for pointers to zero-sized types.
			p.checkIsMethod(tn, f.Type, f.Body)
		}
	}
}

func (p Pass) checkIsMethod(tn *types.TypeName, fType *ast.FuncType, body *ast.BlockStmt) {
	q, ok := p.matchAssertionQuery(fType, body)
	if !ok {
		return
	}

	if atn, _, ok := typeutil.TypeNameOf(types.Unalias(q.assertedType)); !ok || atn != tn {
		return
	}

	// If aliased, use the alias name
	atn, ptr, ok := typeutil.TypeNameOf(q.assertedType)
	if !ok {
		return
	}

	name := shortNameOf(p.Pkg, atn, ptr)
	p.Report(analysis.Diagnostic{
		Pos:     fType.Pos(),
		End:     fType.End(),
		Message: fmt.Sprintf("Is method implementation makes errors.Is act as a type check; remove it and use errors.AsType[%s] at call sites (et:ias)", name),
	})
}

func shortNameOf(current *types.Package, tn *types.TypeName, ptr bool) string {
	name := types.TypeString(tn.Type(), func(pkg *types.Package) string {
		if pkg == current {
			return ""
		}

		return pkg.Name()
	})

	if ptr {
		name = "*" + name
	}

	return name
}
