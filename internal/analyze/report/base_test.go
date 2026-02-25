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

package report_test

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/analyze/report"
)

func checkPackage(tb testing.TB, pkgPath, src string) (*types.Package, ast.Expr) {
	tb.Helper()

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "p.go", src, 0)
	if err != nil {
		tb.Fatalf("ParseFile(%s): %v", pkgPath, err)
	}

	pkg, err := (&types.Config{Importer: importer.Default()}).Check(pkgPath, fset, []*ast.File{file}, nil)
	if err != nil {
		tb.Fatalf("Check(%s): %v", pkgPath, err)
	}

	return pkg, file.Name
}

func TestBase_UndeterminedUsage(t *testing.T) {
	t.Parallel()

	curPkg, expr := checkPackage(t, "example.com/mod/a", `package a
import "fmt"
type MyError struct{ error }
type GenStrError[T ~string] struct{ error }
type GenCustomError[T fmt.Stringer] struct{ error }
`)
	subPkg, _ := checkPackage(t, "example.com/mod/sub", `package sub
type SubError struct{ error }
`)
	extPkg, _ := checkPackage(t, "other.example/lib", `package lib
type ExtError struct{ error }
type extPrivateError struct{ error }
`)
	mod := &analysis.Module{Path: "example.com/mod"}

	tests := [...]struct {
		name        string
		pkg         *types.Package
		typeName    string
		ptr         bool
		wantSub     string
		excludeSub  string
		wantCode    string
		wantRelated bool
	}{
		{
			name:        "same package",
			pkg:         curPkg,
			typeName:    "MyError",
			wantSub:     "var _ error = MyError{}",
			wantCode:    "(et:emb)",
			wantRelated: true,
		},
		{
			name:        "generic satisfiable constraint",
			pkg:         curPkg,
			typeName:    "GenStrError",
			ptr:         true,
			wantSub:     "(*GenStrError[string])(nil)",
			wantCode:    "(et:emb+)",
			wantRelated: true,
		},
		{
			name:        "generic custom constraint",
			pkg:         curPkg,
			typeName:    "GenCustomError",
			wantSub:     "GenCustomError[T]",
			wantCode:    "(et:emb)",
			wantRelated: true,
		},
		{
			name:     "same module",
			pkg:      subPkg,
			typeName: "SubError",
			wantSub:  "example.com/mod/sub",
			wantCode: "(et:emb)",
		},
		{
			name:     "external exported",
			pkg:      extPkg,
			typeName: "ExtError",
			wantSub:  "var _ error = lib.ExtError{}",
			wantCode: "(et:emb)",
		},
		{
			name:       "external unexported",
			pkg:        extPkg,
			typeName:   "extPrivateError",
			wantSub:    "overrides",
			excludeSub: "var _ error = ",
			wantCode:   "(et:emb)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tn := tt.pkg.Scope().Lookup(tt.typeName).(*types.TypeName)

			var got []analysis.Diagnostic

			pass := &analysis.Pass{
				Pkg:    curPkg,
				Module: mod,
				Report: func(d analysis.Diagnostic) { got = append(got, d) },
			}

			report.Base{Pass: pass, Expr: expr}.UndeterminedUsage(tn, tt.ptr)

			if len(got) != 1 {
				t.Fatalf("got %d diagnostics, want 1", len(got))
			}

			msg := got[0].Message
			if !strings.Contains(msg, tt.wantSub) || !strings.Contains(msg, tt.wantCode) {
				t.Errorf("Message %q missing %q or %q", msg, tt.wantSub, tt.wantCode)
			}

			if tt.excludeSub != "" && strings.Contains(msg, tt.excludeSub) {
				t.Errorf("Message %q should not contain %q", msg, tt.excludeSub)
			}

			if tt.wantRelated && (len(got[0].Related) != 1 || got[0].Related[0].Pos != tn.Pos()) {
				t.Errorf("Related = %+v, want single entry at type pos %v", got[0].Related, tn.Pos())
			}

			if !tt.wantRelated && len(got[0].Related) != 0 {
				t.Errorf("Related = %+v, want empty", got[0].Related)
			}
		})
	}
}
