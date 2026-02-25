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

package typeutil_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	. "fillmore-labs.com/errortype/internal/typeutil"
)

func TestPkgPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pkgPath string
		files   []string // file names of the analyzed package
		others  []string // file names of other packages in the same FileSet
		want    string
	}{
		{
			name:    "regular package",
			pkgPath: "example.com/p",
			files:   []string{"p.go"},
			want:    "example.com/p",
		},
		{
			name:    "package with test files",
			pkgPath: "example.com/p",
			files:   []string{"p.go", "p_test.go"},
			want:    "example.com/p (test)",
		},
		{
			name:    "external test package",
			pkgPath: "example.com/p_test",
			files:   []string{"x_test.go"},
			want:    "example.com/p_test (test)",
		},
		{
			name:    "shared FileSet with test files of other packages",
			pkgPath: "example.com/p",
			files:   []string{"p.go"},
			others:  []string{"q.go", "q_test.go"},
			want:    "example.com/p",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fset := token.NewFileSet()

			for _, name := range tt.others {
				parseFile(t, fset, name, "package q")
			}

			files := make([]*ast.File, 0, len(tt.files))
			for _, name := range tt.files {
				files = append(files, parseFile(t, fset, name, "package p"))
			}

			pkg := types.NewPackage(tt.pkgPath, "p")

			if got := PkgPath(fset, pkg, files); got != tt.want {
				t.Errorf("PkgPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func parseFile(tb testing.TB, fset *token.FileSet, name, src string) *ast.File {
	tb.Helper()

	f, err := parser.ParseFile(fset, name, src, parser.SkipObjectResolution)
	if err != nil {
		tb.Fatalf("parse %s: %v", name, err)
	}

	return f
}
