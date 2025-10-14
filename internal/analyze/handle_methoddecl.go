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

var _errorMethods = map[string]func(*types.Signature) bool{
	"Is":     typeutil.HasIsSig,
	"Unwrap": typeutil.HasUnwrapSig,
	"As":     typeutil.HasAsSig,
}

func (p Pass) handleMethodDecl(c inspector.Cursor, n *ast.FuncDecl) {
	sigCheck, ok := _errorMethods[n.Name.Name]
	if !ok {
		return // Not an error related function
	}

	fun, ok := p.TypesInfo.Defs[n.Name].(*types.Func)
	if !ok { // should not happen
		return
	}

	recv := fun.Signature().Recv()
	if recv == nil { // should not happen
		return
	}

	tn, ptr, ok := typeutil.TypeNameOf(recv.Type())
	if !ok { // should not happen
		return
	}

	prop, ok := p.ErrorUsage[tn]
	if !ok {
		return // not an error type
	}

	if !sigCheck(fun.Signature()) {
		// also found by golang.org/x/tools/go/analysis/passes/stdmethods.
		p.ReportRangef(n.Recv, "Method %q has the wrong signature (et:sig)", fun.Name())

		return
	}

	if ptr && prop&usage.ExpectedMask == usage.ValueExpected {
		p.ReportRangef(n.Recv, "Method %q should be implemented with a value receiver, not a pointer (et:rcv)", fun.Name())
	}

	p.checkForIsDecl(c, n)
}

// checkForIsDecl checks for an "Is" method declaration and hands over to body analysis.
func (p Pass) checkForIsDecl(c inspector.Cursor, n *ast.FuncDecl) {
	if n.Name.Name != "Is" || n.Body == nil {
		return
	}

	recv, hasReceiver := singleField(n.Recv)
	if !hasReceiver { // should not happen, checked before
		return
	}

	target, hasParameter := singleField(n.Type.Params)
	if !hasParameter { // should not happen, checked before
		return
	}

	b := c.ChildAt(edge.FuncDecl_Body, -1)
	p.handleIs(b, recv, target)
}

// singleField checks if a field list contains exactly one field, which itself has at most one name.
func singleField(f *ast.FieldList) (name *ast.Ident, ok bool) {
	if f == nil || len(f.List) != 1 || len(f.List[0].Names) > 1 {
		return nil, false
	}

	if field := f.List[0]; len(field.Names) == 1 {
		return field.Names[0], true
	}

	return nil, true
}
