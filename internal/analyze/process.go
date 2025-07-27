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

	"golang.org/x/tools/go/ast/edge"
	"golang.org/x/tools/go/ast/inspector"

	"fillmore-labs.com/errortype/internal/detect"
	"fillmore-labs.com/errortype/internal/errortypes"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// processDetectedTypes populates the initial error usage map based on the results
// from the prerequisite `detecttypes` analyzer.
func (p pass) processDetectedTypes(resultInfo []detect.ResultInfo) {
	for _, detectedType := range resultInfo {
		var usage Usage

		switch detectedType.ErrorType {
		case errortypes.PointerType:
			usage = PointerExpected

		case errortypes.ValueType:
			usage = ValueExpected

		case errortypes.SuppressType:
			usage = SuppressExpected

		default:
			continue
		}

		p.errorUsages[detectedType.TypeName] = usage
	}
}

// processAST traverses the abstract syntax tree of the package being analyzed.
// It visits nodes relevant to error usage and dispatches each to its
// corresponding handler function.
func (p pass) processAST(in *inspector.Inspector, styleCheck bool) {
	for c := range in.Root().Preorder(
		(*ast.CallExpr)(nil),
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
		(*ast.TypeAssertExpr)(nil),
		(*ast.TypeSwitchStmt)(nil),
	) {
		switch n := c.Node().(type) {
		case *ast.CallExpr:
			p.handleErrorsAs(n, styleCheck)

		case *ast.FuncDecl:
			if n.Body == nil {
				continue // Skip function declarations without a body.
			}

			if lastResult := typeutil.HasErrorResult(p.TypesInfo, n.Type.Results); lastResult >= 0 {
				b := c.ChildAt(edge.FuncDecl_Body, -1)
				p.handleReturns(b, lastResult)
			}

		case *ast.FuncLit:
			if lastResult := typeutil.HasErrorResult(p.TypesInfo, n.Type.Results); lastResult >= 0 {
				b := c.ChildAt(edge.FuncLit_Body, -1)
				p.handleReturns(b, lastResult)
			}

		case *ast.TypeAssertExpr:
			p.handleTypeAssert(n)

		case *ast.TypeSwitchStmt:
			p.handleTypeSwitch(n)
		}
	}
}
