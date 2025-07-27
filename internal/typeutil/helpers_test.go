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

package typeutil_test

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"
)

// parseSource is a helper to parse source and get [types.Info] and [types.Package].
func parseSource(tb testing.TB, src string) (*types.Info, *types.Package, *token.FileSet, *ast.File) { //nolint:unparam
	tb.Helper()

	const (
		filename = "test.go"
		pkgname  = "test"
	)

	var sb strings.Builder
	sb.WriteString("package ")
	sb.WriteString(pkgname)
	sb.WriteString("\n")
	sb.WriteString(src)

	srcFile := sb.String()

	fset := token.NewFileSet()
	fset.AddFile(filename, -1, len(srcFile))

	f, err := parser.ParseFile(fset, filename, srcFile, parser.SkipObjectResolution)
	if err != nil {
		tb.Fatalf("failed to parse source: %T %v", err, err)
	}

	conf := types.Config{Importer: importer.Default()}

	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	pkg, err := conf.Check(pkgname, fset, []*ast.File{f}, info)
	if err != nil {
		tb.Fatalf("failed to type Check source: %v", err)
	}

	return info, pkg, fset, f
}
