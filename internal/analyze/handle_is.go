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
	"strings"

	"golang.org/x/tools/go/ast/inspector"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// handleIs identifies calls of errors.Is inside Is(error) methods.
func (p pass) handleIs(b inspector.Cursor, recv, target *ast.Ident, deepIsCheck bool) {
	targetVar, ok := p.TypesInfo.Defs[target].(*types.Var)
	if !ok { // should not happen
		p.ReportErrorf(b.Node(), "Can't determine parameter %v", target)
		return
	}

	for c := range b.Preorder((*ast.CallExpr)(nil)) {
		node, ok := c.Node().(*ast.CallExpr)
		if !ok { // should not happen
			continue
		}

		fun, _, methodExpr, ok := typeutil.FuncOf(p.TypesInfo, node.Fun)
		if !ok {
			continue // Could not resolve the calls function, might be a func variable.
		}

		funcName := typeutil.FuncNameOf(fun)

		info, ok := typeutil.KnownFuncs[funcName]
		if !ok {
			continue
		}

		var errArgIndex int

		switch info.Kind() {
		case typeutil.KindIs:
			switch info.IsType() {
			case typeutil.IsFunc0:
				errArgIndex = 0

			case typeutil.IsFunc1:
				errArgIndex = 1

			default:
				continue
			}

		case typeutil.KindEqu:
			continue

		case typeutil.KindAs:
			errArgIndex, _ = info.AsTarget()
			errArgIndex--
		}

		if errArgIndex < 0 {
			continue
		}

		if methodExpr {
			errArgIndex++
		}

		if len(node.Args) <= errArgIndex {
			continue
		}

		arg := node.Args[errArgIndex]
		id, isIdent := ast.Unparen(arg).(*ast.Ident)

		codeSuffix := ""
		if isIdent && p.isTargetArg(id, target, targetVar) {
			codeSuffix = "+"
		} else if !deepIsCheck {
			continue
		}

		names := buildNames(recv, target)
		p.ReportRangef(node, "An Is method should only shallowly compare %s. (et:unw%s)", names, codeSuffix)
	}
}

// isTargetArg determines whether arg is an identifier that refers to the target variable.
func (p pass) isTargetArg(id, target *ast.Ident, targetVar *types.Var) bool {
	// arg resolves to the target variable.
	return id.Name == target.Name && p.TypesInfo.Uses[id] == targetVar
}

// buildNames constructs a string like "v and target" from the receiver and parameter identifiers.
func buildNames(receiver, param *ast.Ident) string {
	var names strings.Builder
	if receiver != nil && receiver.Name != "" && receiver.Name != "_" {
		names.WriteString(receiver.Name)
	}

	if param != nil && param.Name != "" && param.Name != "_" {
		if names.Len() != 0 {
			names.WriteString(" and ")
		}

		names.WriteString(param.Name)
	}

	return names.String()
}
