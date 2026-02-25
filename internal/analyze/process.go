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
	"context"
	"go/ast"
	"runtime/trace"

	"golang.org/x/tools/go/ast/edge"
	"golang.org/x/tools/go/ast/inspector"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// ProcessAST traverses the abstract syntax tree of the package being analyzed.
// It visits nodes relevant to error usage and dispatches each to its
// corresponding handler function.
func (p Pass) ProcessAST(ctx context.Context, in *inspector.Inspector) {
	defer trace.StartRegion(ctx, "AST").End()

	root := in.Root()
	for c := range root.Preorder(
		// keep-sorted start
		(*ast.AssignStmt)(nil),
		(*ast.BinaryExpr)(nil),
		(*ast.CallExpr)(nil),
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
		(*ast.TypeAssertExpr)(nil),
		(*ast.TypeSwitchStmt)(nil),
		(*ast.ValueSpec)(nil),
		// keep-sorted end
	) {
		var reg *trace.Region

		switch n := c.Node().(type) {
		// keep-sorted start newline_separated=yes
		case *ast.AssignStmt:
			reg = trace.StartRegion(ctx, "AssignStmt")

			p.handleAssign(n)

		case *ast.BinaryExpr:
			reg = trace.StartRegion(ctx, "BinaryExpr")

			p.handleBinaryExpr(n)

		case *ast.CallExpr:
			reg = trace.StartRegion(ctx, "CallExpr")

			p.handleCall(c, n)

		case *ast.FuncDecl:
			reg = trace.StartRegion(ctx, "FuncDecl")

			if n.Recv != nil {
				p.handleMethodDecl(c, n)
			}

			if n.Body == nil {
				break // Skip function declarations without a body.
			}

			if errResultIdx := typeutil.ErrorResultIndex(p.TypesInfo, n.Type.Results); errResultIdx >= 0 {
				body := c.ChildAt(edge.FuncDecl_Body, -1)
				p.handleReturns(body, errResultIdx)
			}

		case *ast.FuncLit:
			reg = trace.StartRegion(ctx, "FuncLit")

			if errResultIdx := typeutil.ErrorResultIndex(p.TypesInfo, n.Type.Results); errResultIdx >= 0 {
				body := c.ChildAt(edge.FuncLit_Body, -1)
				p.handleReturns(body, errResultIdx)
			}

		case *ast.TypeAssertExpr:
			reg = trace.StartRegion(ctx, "TypeAssert")

			p.handleTypeAssert(n)

		case *ast.TypeSwitchStmt:
			reg = trace.StartRegion(ctx, "TypeSwitch")

			p.handleTypeSwitch(n)

		case *ast.ValueSpec:
			reg = trace.StartRegion(ctx, "ValueSpec")

			p.handleVarDecls(n)
			// keep-sorted end
		}

		if reg != nil {
			reg.End()
		}
	}
}
