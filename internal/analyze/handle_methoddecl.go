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

	"golang.org/x/tools/go/ast/edge"
	"golang.org/x/tools/go/ast/inspector"

	"fillmore-labs.com/errortype/internal/analyze/usage"
	"fillmore-labs.com/errortype/internal/typeutil"
)

func (p Pass) handleMethodDecl(c inspector.Cursor, f *ast.FuncDecl) {
	var sigCheck func(*types.Signature) bool

	switch f.Name.Name {
	case "Is":
		sigCheck = typeutil.HasIsSig

	case "Unwrap":
		sigCheck = typeutil.HasUnwrapSig

	case "As":
		sigCheck = typeutil.HasAsSig

	default:
		return // Not an error-related function
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

	if !sigCheck(sig) {
		// also found by golang.org/x/tools/go/analysis/passes/stdmethods.
		p.ReportRangef(f.Type, "Method %q has the wrong signature (et:sig)", fun.Name())

		return
	}

	if ptr && prop&usage.ExpectedMask == usage.ValueExpected {
		p.ReportRangef(f.Recv, "Method %q should be implemented with a value receiver, not a pointer (et:rcv)", fun.Name())
	}

	// check for an "Is" method declaration and hand over to body analysis.
	if f.Name.Name == "Is" && f.Body != nil {
		param := singleField(f.Type.Params)
		if param == nil || param.Name == "_" {
			return // An "Is" method without named target parameter is legal, but rare.
		}

		target, ok := p.TypesInfo.Defs[param].(*types.Var)
		if !ok { // should not happen
			p.ReportErrorf(param, "Can't determine parameter type of %q", param.Name)
			return
		}

		body := c.ChildAt(edge.FuncDecl_Body, -1)
		p.handleIs(body, target)
	}
}

// singleField checks if a field list contains exactly one field, which itself has at most one name.
func singleField(f *ast.FieldList) *ast.Ident {
	if f == nil || len(f.List) != 1 {
		return nil
	}

	field := f.List[0]
	if len(field.Names) != 1 {
		return nil
	}

	return field.Names[0]
}
