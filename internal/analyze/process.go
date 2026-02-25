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
	"fmt"
	"go/ast"
	"runtime/trace"

	"golang.org/x/tools/go/ast/edge"
	"golang.org/x/tools/go/ast/inspector"

	"fillmore-labs.com/errortype/internal/astutil"
	"fillmore-labs.com/errortype/internal/naming/diagnostic"
	"fillmore-labs.com/errortype/internal/naming/rules"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// ProcessAST traverses the abstract syntax tree of the package being analyzed.
// It visits nodes relevant to error usage and dispatches each to its
// corresponding handler function.
func (p Pass) ProcessAST(ctx context.Context, in *inspector.Inspector, fileMap astutil.FileMap) error {
	defer trace.StartRegion(ctx, "AST").End()

	var naming *rules.Checker
	if p.Has(OptionNaming) {
		naming = rules.New(p.TypesInfo, p.Analyzer.Name)
	}

	root := in.Root()
	for c, cok := root.FirstChild(); cok; c, cok = c.NextSibling() {
		file := c.Node().(*ast.File)
		// Skip generated files
		if !p.Has(OptionGenerated) && fileMap.IsGenerated(file) {
			continue
		}

		// Skip files with nolint comment
		if astutil.HasNoLint(file.Doc, p.Analyzer.Name) {
			continue
		}

		for c := range c.Preorder(
			// keep-sorted start
			(*ast.AssignStmt)(nil),
			(*ast.BinaryExpr)(nil),
			(*ast.CallExpr)(nil),
			(*ast.FuncDecl)(nil),
			(*ast.FuncLit)(nil),
			(*ast.TypeAssertExpr)(nil),
			(*ast.TypeSpec)(nil),
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

				if n.Recv == nil {
					p.handleFuncDecl(n, file)
				} else {
					p.handleMethodDecl(c, n)
				}

				if n.Body == nil {
					break // Skip function declarations without a body.
				}

				if errResultIdx := typeutil.ErrorResultIndex(p.TypesInfo, n.Type); errResultIdx >= 0 {
					body := c.ChildAt(edge.FuncDecl_Body, -1)
					p.handleReturns(body, errResultIdx)
				}

			case *ast.FuncLit:
				reg = trace.StartRegion(ctx, "FuncLit")

				if errResultIdx := typeutil.ErrorResultIndex(p.TypesInfo, n.Type); errResultIdx >= 0 {
					body := c.ChildAt(edge.FuncLit_Body, -1)
					p.handleReturns(body, errResultIdx)
				}

			case *ast.TypeAssertExpr:
				reg = trace.StartRegion(ctx, "TypeAssert")

				p.handleTypeAssert(n)

			case *ast.TypeSpec:
				reg = trace.StartRegion(ctx, "TypeSpec")

				parent := c.Parent()
				declNode := parent.Node()

				decl, ok := declNode.(*ast.GenDecl)
				if !ok {
					return fmt.Errorf("internal error: unexpected parent of type spec: %#v", declNode)
				}

				// Package level TypeSpec?
				if _, ok := parent.Parent().Node().(*ast.File); ok {
					if naming != nil {
						// Check naming of error types
						naming.CheckTypeSpec(n, decl)
					}

					if ok && p.Has(OptionNotComparable) {
						p.handleTypeSpec(n, decl)
					}
				}

			case *ast.TypeSwitchStmt:
				reg = trace.StartRegion(ctx, "TypeSwitch")

				p.handleTypeSwitch(n)

			case *ast.ValueSpec:
				reg = trace.StartRegion(ctx, "ValueSpec")

				parent := c.Parent()

				decl, ok := parent.Node().(*ast.GenDecl)
				if !ok {
					return fmt.Errorf("internal error: unexpected parent of value spec: %#v", parent.Node())
				}

				_, packageLevel := parent.Parent().Node().(*ast.File)
				if packageLevel && naming != nil {
					// Check naming of error sentinels
					naming.CheckValueSpec(n, decl)
				}

				p.handleValueSpec(n, packageLevel)
				// keep-sorted end
			}

			if reg != nil {
				reg.End()
			}
		}
	}

	if naming != nil {
		violations := naming.Violations()
		if err := diagnostic.ReportViolations(p.Pass, fileMap, violations, namingMessage); err != nil {
			return err
		}
	}

	return nil
}
