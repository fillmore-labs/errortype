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

package result_test

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"fillmore-labs.com/errortype/internal/detectwrappers/result"
	"fillmore-labs.com/errortype/internal/typeutil"
)

func checkCallTestSource(tb testing.TB) (*types.Package, *types.Info, map[string]*ast.CallExpr) {
	tb.Helper()

	const src = `package test

type MyErr struct{}

func (MyErr) Error() string { return "" }

type PtrMyErr = *MyErr

type Suite struct{}

func (Suite) Is(err, target error) bool { return false }

type G[T any] struct{}

func (G[T]) M() {}

func myIs(err, target error) bool            { return false }
func myAs(err error, target any) bool        { return false }
func myAsType[E error](err error) bool       { return false }
func myAssert[T any](v any) bool             { return false }
func myErrorf(fmt string, args ...any) error { return nil }
func other(err error) bool                   { return false }

var (
	err      error
	target   *MyErr
	ptrAlias PtrMyErr

	cDirect     = myIs(err, &MyErr{})
	cMethodExpr = Suite.Is(Suite{}, err, &MyErr{})
	cUnknown    = other(err)
	cAsPtr      = myAs(err, &target)
	cAsPtrAlias = myAs(err, ptrAlias)
	cAsNonPtr   = myAs(err, err)
	cAsType     = myAsType[*MyErr](err)
	cAssert     = myAssert[MyErr](err)
	cErrorf     = myErrorf("msg %w", err)
	_           = G[int]{}.M
)

func localCall(err error) {
	localFn := func(error, error) bool { return false }
	cLocal := localFn(err, &MyErr{})
	_ = cLocal
}
`

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, "test.go", src, parser.SkipObjectResolution)
	if err != nil {
		tb.Fatalf("ParseFile failed: %v", err)
	}

	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
		Instances:  make(map[*ast.Ident]types.Instance),
	}

	conf := types.Config{Importer: importer.Default()}

	pkg, err := conf.Check("test", fset, []*ast.File{f}, info)
	if err != nil {
		tb.Fatalf("Check failed: %v", err)
	}

	calls := make(map[string]*ast.CallExpr)

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok || len(vs.Names) == 0 || len(vs.Values) == 0 {
					continue
				}

				if call, ok := vs.Values[0].(*ast.CallExpr); ok {
					calls[vs.Names[0].Name] = call
				}
			}

		case *ast.FuncDecl:
			if d.Name.Name != "localCall" || d.Body == nil {
				continue
			}

			for _, stmt := range d.Body.List {
				asgn, ok := stmt.(*ast.AssignStmt)
				if !ok || len(asgn.Lhs) == 0 || len(asgn.Rhs) == 0 {
					continue
				}

				id, ok := asgn.Lhs[0].(*ast.Ident)
				if !ok {
					continue
				}

				if call, ok := asgn.Rhs[0].(*ast.CallExpr); ok {
					calls[id.Name] = call
				}
			}
		}
	}

	return pkg, info, calls
}

func TestErrorFuncs_Lookup(t *testing.T) {
	t.Parallel()

	pkg, info, calls := checkCallTestSource(t)

	myIsObj := pkg.Scope().Lookup("myIs")
	suiteIsObj, _, _ := types.LookupFieldOrMethod(pkg.Scope().Lookup("Suite").Type(), false, pkg, "Is")

	efDirect := result.ErrorFunc{Type: result.WrapperIs, Source: 0, Target: 1}
	efMethod := result.ErrorFunc{Type: result.WrapperIs, Source: 0, Target: 1}

	errorFuncs := result.ErrorFuncs{
		myIsObj:    efDirect,
		suiteIsObj: efMethod,
	}

	tests := [...]struct {
		name           string
		call           string
		wantOK         bool
		wantMethodExpr bool
		wantEF         result.ErrorFunc
	}{
		{"direct", "cDirect", true, false, efDirect},
		{"methodExpr", "cMethodExpr", true, true, efMethod},
		{"unknown", "cUnknown", false, false, result.ErrorFunc{}},
		{"localVar", "cLocal", false, false, result.ErrorFunc{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			call, ok := calls[tt.call]
			if !ok {
				t.Fatalf("call %q not found", tt.call)
			}

			fun, ef, gotOK := errorFuncs.Lookup(info, call)
			if gotOK != tt.wantOK {
				t.Fatalf("Lookup(%s) ok = %t, want %t", tt.call, gotOK, tt.wantOK)
			}

			if !gotOK {
				return
			}

			if fun.MethodExpr != tt.wantMethodExpr {
				t.Errorf("Lookup(%s) MethodExpr = %t, want %t", tt.call, fun.MethodExpr, tt.wantMethodExpr)
			}

			if ef != tt.wantEF {
				t.Errorf("Lookup(%s) ErrorFunc = %v, want %v", tt.call, ef, tt.wantEF)
			}
		})
	}
}

func TestErrorFunc_Args(t *testing.T) {
	t.Parallel()

	a0 := &ast.Ident{Name: "a0"}
	a1 := &ast.Ident{Name: "a1"}
	a2 := &ast.Ident{Name: "a2"}
	args := []ast.Expr{a0, a1, a2}

	t.Run("SourceArg", func(t *testing.T) {
		t.Parallel()

		tests := [...]struct {
			name       string
			ef         result.ErrorFunc
			args       []ast.Expr
			methodExpr bool
			want       ast.Expr
			wantOK     bool
		}{
			{"normal", result.ErrorFunc{Type: result.WrapperIs, Source: 0, Target: 1}, args, false, a0, true},
			{"methodExpr", result.ErrorFunc{Type: result.WrapperIs, Source: 0, Target: 1}, args, true, a1, true},
			{"funcAssert", result.ErrorFunc{Type: result.FuncAssert, Source: -1, Target: 0}, args, false, nil, false},
			{"funcAssertMethodExpr", result.ErrorFunc{Type: result.FuncAssert, Source: -1, Target: 0}, args, true, nil, false},
			{"outOfBounds", result.ErrorFunc{Type: result.WrapperIs, Source: 2, Target: 3}, []ast.Expr{a0}, false, nil, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				if got, ok := tt.ef.SourceArg(tt.args, tt.methodExpr); ok != tt.wantOK || got != tt.want {
					t.Errorf("SourceArg() = (%v, %t), want (%v, %t)", got, ok, tt.want, tt.wantOK)
				}
			})
		}
	})

	t.Run("TargetArg", func(t *testing.T) {
		t.Parallel()

		tests := [...]struct {
			name       string
			ef         result.ErrorFunc
			args       []ast.Expr
			methodExpr bool
			want       ast.Expr
			wantOK     bool
		}{
			{"normal", result.ErrorFunc{Type: result.WrapperIs, Source: 0, Target: 1}, args, false, a1, true},
			{"methodExpr", result.ErrorFunc{Type: result.WrapperIs, Source: 0, Target: 1}, args, true, a2, true},
			{"wrapperAsType", result.ErrorFunc{Type: result.WrapperAsType, Source: 0, Target: 0}, args, false, nil, false},
			{"funcAssert", result.ErrorFunc{Type: result.FuncAssert, Source: -1, Target: 0}, args, false, nil, false},
			{"outOfBounds", result.ErrorFunc{Type: result.WrapperAs, Source: 0, Target: 1}, []ast.Expr{a0}, false, nil, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				if got, ok := tt.ef.TargetArg(tt.args, tt.methodExpr); ok != tt.wantOK || got != tt.want {
					t.Errorf("TargetArg() = (%v, %t), want (%v, %t)", got, ok, tt.want, tt.wantOK)
				}
			})
		}
	})

	t.Run("TargetArgs", func(t *testing.T) {
		t.Parallel()

		tests := [...]struct {
			name       string
			ef         result.ErrorFunc
			args       []ast.Expr
			methodExpr bool
			wantLen    int
			wantFirst  ast.Expr
		}{
			{"withOperands", result.ErrorFunc{Type: result.WrapperErrorf, Source: 0, Target: 1}, args, false, 2, a1},
			{"methodExpr", result.ErrorFunc{Type: result.WrapperErrorf, Source: 0, Target: 1}, args, true, 1, a2},
			{"noOperands", result.ErrorFunc{Type: result.WrapperErrorf, Source: 0, Target: 1}, []ast.Expr{a0}, false, 0, nil},
			{"nonErrorf", result.ErrorFunc{Type: result.WrapperIs, Source: 0, Target: 1}, args, false, 0, nil},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got := tt.ef.TargetArgs(tt.args, tt.methodExpr)
				if len(got) != tt.wantLen {
					t.Fatalf("TargetArgs() len = %d, want %d", len(got), tt.wantLen)
				}

				if tt.wantLen > 0 && got[0] != tt.wantFirst {
					t.Errorf("TargetArgs()[0] = %v, want %v", got[0], tt.wantFirst)
				}
			})
		}
	})
}

func TestErrorFunc_TargetType(t *testing.T) {
	t.Parallel()

	pkg, info, calls := checkCallTestSource(t)
	myErrType := pkg.Scope().Lookup("MyErr").Type()
	ptrMyErrType := types.NewPointer(myErrType)

	tests := [...]struct {
		name     string
		call     string
		ef       result.ErrorFunc
		wantType types.Type
		wantOK   bool
	}{
		{
			name:     "WrapperIs",
			call:     "cDirect",
			ef:       result.ErrorFunc{Type: result.WrapperIs, Source: 0, Target: 1},
			wantType: ptrMyErrType,
			wantOK:   true,
		},
		{
			name:     "WrapperAs_pointer",
			call:     "cAsPtr",
			ef:       result.ErrorFunc{Type: result.WrapperAs, Source: 0, Target: 1},
			wantType: ptrMyErrType,
			wantOK:   true,
		},
		{
			name:     "WrapperAs_pointerAlias",
			call:     "cAsPtrAlias",
			ef:       result.ErrorFunc{Type: result.WrapperAs, Source: 0, Target: 1},
			wantType: myErrType,
			wantOK:   true,
		},
		{
			name:   "WrapperAs_nonPointer",
			call:   "cAsNonPtr",
			ef:     result.ErrorFunc{Type: result.WrapperAs, Source: 0, Target: 1},
			wantOK: false,
		},
		{
			name:     "WrapperAsType",
			call:     "cAsType",
			ef:       result.ErrorFunc{Type: result.WrapperAsType, Source: 0, Target: 0},
			wantType: ptrMyErrType,
			wantOK:   true,
		},
		{
			name:   "WrapperAsType_outOfBounds",
			call:   "cAsType",
			ef:     result.ErrorFunc{Type: result.WrapperAsType, Source: 0, Target: 1},
			wantOK: false,
		},
		{
			name:     "FuncAssert",
			call:     "cAssert",
			ef:       result.ErrorFunc{Type: result.FuncAssert, Source: -1, Target: 0},
			wantType: myErrType,
			wantOK:   true,
		},
		{
			name:   "WrapperErrorf",
			call:   "cErrorf",
			ef:     result.ErrorFunc{Type: result.WrapperErrorf, Source: 0, Target: 1},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			call, ok := calls[tt.call]
			if !ok {
				t.Fatalf("call %q not found", tt.call)
			}

			fun, ok := typeutil.FuncOf(info, call)
			if !ok {
				t.Fatalf("FuncOf(%s) not ok", tt.call)
			}

			got, gotOK := tt.ef.TargetType(info, fun, call.Args)
			if gotOK != tt.wantOK {
				t.Fatalf("TargetType(%s) = (%v, %t), want ok = %t", tt.call, got, gotOK, tt.wantOK)
			}

			if !gotOK {
				return
			}

			if !types.Identical(got, tt.wantType) {
				t.Errorf("TargetType(%s) = %s, want %s", tt.call, got, tt.wantType)
			}
		})
	}
}

func TestErrorFuncs_IsAsTypeTarget(t *testing.T) {
	t.Parallel()

	pkg, info, _ := checkCallTestSource(t)

	myAsTypeObj := pkg.Scope().Lookup("myAsType")
	myIsObj := pkg.Scope().Lookup("myIs")
	myAssertObj := pkg.Scope().Lookup("myAssert")
	otherObj := pkg.Scope().Lookup("other")
	gOriginObj, _, _ := types.LookupFieldOrMethod(pkg.Scope().Lookup("G").Type(), false, pkg, "M")

	var gInstMethod *types.Func

	for id, obj := range info.Uses {
		if id.Name != "M" {
			continue
		}

		if fn, ok := obj.(*types.Func); ok && fn.Origin() == gOriginObj && fn != gOriginObj {
			gInstMethod = fn
			break
		}
	}

	if gInstMethod == nil {
		t.Fatal("instantiated method G[int].M not found in info.Uses")
	}

	errorFuncs := result.ErrorFuncs{
		myAsTypeObj: {Type: result.WrapperAsType, Source: 0, Target: 1},
		myIsObj:     {Type: result.WrapperIs, Source: 0, Target: 1},
		myAssertObj: {Type: result.FuncAssert, Source: -1, Target: 0},
		gOriginObj:  {Type: result.WrapperAsType, Source: 0, Target: 0},
	}

	tests := [...]struct {
		name  string
		obj   types.Object
		index int
		want  bool
	}{
		{"matching target index", myAsTypeObj, 1, true},
		{"non-matching target index", myAsTypeObj, 0, false},
		{"wrapper is", myIsObj, 1, false},
		{"func assert", myAssertObj, 0, false},
		{"unknown object", otherObj, 0, false},
		{"nil object", nil, 0, false},
		{"instantiated method via origin", gInstMethod, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := errorFuncs.IsAsTypeTarget(tt.obj, tt.index); got != tt.want {
				t.Errorf("IsAsTypeTarget(%v, %d) = %t, want %t", tt.obj, tt.index, got, tt.want)
			}
		})
	}
}
