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

	"fillmore-labs.com/errortype/internal/errortypes"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// processDetectedTypes populates the initial error usage map based on the results
// from the prerequisite `detecttypes` analyzer.
func (p pass) processDetectedTypes(ctx context.Context, resultInfo []errortypes.ResultInfo) {
	defer trace.StartRegion(ctx, "detectedTypes").End()

	for _, detectedType := range resultInfo {
		var usage Usage

		switch detectedType.ErrorType & errortypes.ExpectedMask {
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
func (p pass) processAST(ctx context.Context, in *inspector.Inspector, opts AstOptions) {
	defer trace.StartRegion(ctx, "AST").End()

	for c := range in.Root().Preorder(
		// keep-sorted start
		(*ast.BinaryExpr)(nil),
		(*ast.CallExpr)(nil),
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
		(*ast.TypeAssertExpr)(nil),
		(*ast.TypeSwitchStmt)(nil),
		(*ast.ValueSpec)(nil),
		// keep-sorted end

	) {
		switch n := c.Node().(type) {
		// keep-sorted start newline_separated=yes
		case *ast.BinaryExpr:
			reg := trace.StartRegion(ctx, "BinaryExpr")

			p.handleBinaryExpr(n)
			reg.End()

		case *ast.CallExpr:
			reg := trace.StartRegion(ctx, "CallExpr")

			p.handleCall(n, c, opts)
			reg.End()

		case *ast.FuncDecl:
			reg := trace.StartRegion(ctx, "FuncDecl")

			if n.Recv != nil {
				p.handleMethodDecl(n, c, opts)
			}

			if n.Body == nil {
				reg.End()
				continue // Skip function declarations without a body.
			}

			if lastResult := typeutil.HasErrorResult(p.TypesInfo, n.Type.Results); lastResult >= 0 {
				b := c.ChildAt(edge.FuncDecl_Body, -1)
				p.handleReturns(b, lastResult)
			}

			reg.End()

		case *ast.FuncLit:
			reg := trace.StartRegion(ctx, "FuncLit")

			if lastResult := typeutil.HasErrorResult(p.TypesInfo, n.Type.Results); lastResult >= 0 {
				b := c.ChildAt(edge.FuncLit_Body, -1)
				p.handleReturns(b, lastResult)
			}

			reg.End()

		case *ast.TypeAssertExpr:
			reg := trace.StartRegion(ctx, "TypeAssert")

			p.handleTypeAssert(n, opts.UncheckedAssert)
			reg.End()

		case *ast.TypeSwitchStmt:
			reg := trace.StartRegion(ctx, "TypeSwitch")

			p.handleTypeSwitch(n)
			reg.End()

		case *ast.ValueSpec:
			reg := trace.StartRegion(ctx, "ValueSpec")

			p.handleVarDecls(n)
			reg.End()
			// keep-sorted end
		}
	}
}
