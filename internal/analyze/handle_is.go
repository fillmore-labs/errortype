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
func (p pass) handleIs(b inspector.Cursor, err, target *ast.Ident, deepIsCheck bool) {
	targetVar, ok := p.TypesInfo.Defs[target].(*types.Var)
	if !ok { // should not happen
		p.ReportErrorf(b.Node(), "Can't determine parameter %v", target)

		return
	}

	for c := range b.Preorder((*ast.CallExpr)(nil)) {
		node, ok := c.Node().(*ast.CallExpr)
		if !ok {
			continue
		}

		fun, _, methodExpr, ok := typeutil.FuncOf(p.TypesInfo, node.Fun)
		if !ok {
			continue // Could not resolve function, might be a func variable.
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

		plus := ""
		if p.isTargetArg(node.Args[errArgIndex], target, targetVar) {
			plus = "+"
		} else if !deepIsCheck {
			continue
		}

		var names strings.Builder
		if err != nil && err.Name != "" && err.Name != "_" {
			names.WriteString(err.Name)
		}

		if target != nil && target.Name != "" && target.Name != "_" {
			if names.Len() != 0 {
				names.WriteString(" and ")
			}

			names.WriteString(target.Name)
		}

		p.ReportRangef(node, "An Is method should only shallowly compare %s (et:unw%s)", names.String(), plus)
	}
}

// isTargetArg determindes whether arg is an identifier that refers to the target variable.
func (p pass) isTargetArg(arg ast.Expr, target *ast.Ident, targetVar *types.Var) bool {
	id, ok := ast.Unparen(arg).(*ast.Ident)
	if !ok || id.Name != target.Name {
		return false // no an identifier or the wrong one
	}

	// arg resolves to the target variable.
	return p.TypesInfo.Uses[id] == targetVar
}

// isIs determindes whether the function is an `Is(error)bool` method on an error type.
func (p pass) isIs(n *ast.FuncDecl) (err, target *ast.Ident, ok bool) {
	if n.Name.Name != "Is" {
		return nil, nil, false
	}

	err, recv, ok := singleParam(n.Recv)
	if !ok {
		return nil, nil, false
	}

	recvTV := p.TypesInfo.Types[recv]

	tn, _, ok := typeutil.TypeNameOf(recvTV.Type)
	if !ok {
		return nil, nil, false
	}

	if _, ok = p.errorUsages.GetTypeProperty(tn); !ok {
		return nil, nil, false
	}

	target, param, ok := singleParam(n.Type.Params)
	if !ok {
		return nil, nil, false
	}

	_, result, ok := singleParam(n.Type.Results)
	if !ok {
		return nil, nil, false
	}

	argTV := p.TypesInfo.Types[param]
	argType := types.Unalias(argTV.Type)

	if !types.Identical(argType, universeError) {
		return nil, nil, false // Wrong result type
	}

	resultTV := p.TypesInfo.Types[result]
	if b, basic := types.Unalias(resultTV.Type).(*types.Basic); !basic || b.Kind() != types.Bool {
		return nil, nil, false // Wrong result type
	}

	return err, target, true
}

func singleParam(f *ast.FieldList) (name *ast.Ident, typ ast.Expr, ok bool) {
	if f == nil || len(f.List) != 1 || len(f.List[0].Names) > 1 {
		return nil, nil, false
	}

	switch len(f.List[0].Names) {
	case 0:
		return nil, f.List[0].Type, true

	case 1:
		return f.List[0].Names[0], f.List[0].Type, true

	default:
		return nil, nil, false
	}
}
