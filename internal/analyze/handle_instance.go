// Copyright 2026 Oliver Eikemeier. All Rights Reserved.
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
	"cmp"
	"context"
	"go/ast"
	"runtime/trace"
	"slices"

	"golang.org/x/tools/go/ast/edge"
	"golang.org/x/tools/go/ast/inspector"

	"fillmore-labs.com/errortype/internal/astutil"
	"fillmore-labs.com/errortype/internal/errorconv"
)

// groupInstanceTypeArgs collects generic instantiations whose type parameter
// constraint has an `Error() string` method, sorts them by source position,
// and slices contiguous runs per file.
func (p Pass) groupInstanceTypeArgs(fileMap astutil.FileMap) map[*ast.File][]errorconv.TypeArg {
	if len(p.TypesInfo.Instances) == 0 {
		return nil
	}

	args := slices.Collect(errorconv.TypeArgs(p.TypesInfo))
	if len(args) == 0 {
		return nil
	}

	slices.SortFunc(args, func(a, b errorconv.TypeArg) int {
		return cmp.Or(cmp.Compare(a.Ident.Pos(), b.Ident.Pos()), cmp.Compare(a.Index, b.Index))
	})

	byFile := make(map[*ast.File][]errorconv.TypeArg)

	for i := 0; i < len(args); {
		file := fileMap.File(args[i].Ident.Pos())
		if file == nil {
			i++

			continue
		}

		end := i + 1
		for end < len(args) && args[end].Ident.Pos() <= file.FileEnd {
			end++
		}

		byFile[file] = args[i:end]
		i = end
	}

	return byFile
}

// handleFileInstances checks sorted generic instantiations within a single file
// where a type parameter constraint has an `Error() string` method, handling both
// AsType wrapper targets and general generic instantiations without double diagnosis.
func (p Pass) handleFileInstances(ctx context.Context, fileCursor inspector.Cursor, args []errorconv.TypeArg) {
	defer trace.StartRegion(ctx, "Instances").End()

	var (
		prevID *ast.Ident
		ictx   instanceContext
		hasCtx bool
	)

	for _, arg := range args {
		if arg.Ident != prevID {
			prevID = arg.Ident
			ictx, hasCtx = resolveInstanceContext(fileCursor, arg.Ident)
		}

		if !hasCtx {
			continue
		}

		if p.IsAsTypeTarget(arg.Obj, arg.Index) {
			reportExpr := ictx.instExpr
			if ictx.callExpr != nil {
				reportExpr = ictx.callExpr
			}

			p.checkGenericCall(arg.Type, reportExpr, ictx.funExpr)

			continue
		}

		p.checkInstance(arg.Type, ictx.reportExpr(arg.Index))
	}
}

type instanceContext struct {
	funExpr  ast.Expr
	instExpr ast.Expr
	indices  []ast.Expr
	callExpr *ast.CallExpr
}

func (c instanceContext) reportExpr(index int) ast.Expr {
	if index < len(c.indices) {
		return c.indices[index]
	}

	if c.callExpr != nil {
		return c.callExpr
	}

	return c.instExpr
}

func resolveInstanceContext(fileCursor inspector.Cursor, id *ast.Ident) (instanceContext, bool) {
	c, ok := fileCursor.FindNode(id)
	if !ok {
		return instanceContext{}, false
	}

	if c.ParentEdgeKind() == edge.SelectorExpr_Sel {
		c = c.Parent()
	}

	funExpr := c.Node().(ast.Expr)
	outer := c.UnparenEnclosing()

	var indices []ast.Expr

	switch outer.ParentEdgeKind() {
	case edge.IndexExpr_X:
		outer = outer.Parent()
		indices = []ast.Expr{outer.Node().(*ast.IndexExpr).Index}

	case edge.IndexListExpr_X:
		outer = outer.Parent()
		indices = outer.Node().(*ast.IndexListExpr).Indices
	}

	instExpr := outer.Node().(ast.Expr)
	outer = outer.UnparenEnclosing()

	var callExpr *ast.CallExpr
	if outer.ParentEdgeKind() == edge.CallExpr_Fun {
		callExpr = outer.Parent().Node().(*ast.CallExpr)
	}

	return instanceContext{
		funExpr:  funExpr,
		instExpr: instExpr,
		indices:  indices,
		callExpr: callExpr,
	}, true
}
