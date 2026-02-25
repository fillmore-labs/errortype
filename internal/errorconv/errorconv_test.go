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

package errorconv_test

import (
	"cmp"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"slices"
	"testing"

	"fillmore-labs.com/errortype/internal/errorconv"
)

type wantConv struct {
	typ  string
	expr string
}

const commonDecls = `package p

type ValErr struct{}
func (ValErr) Error() string { return "val" }

type PtrErr struct{}
func (*PtrErr) Error() string { return "ptr" }

type CustomErr interface {
	error
	Code() int
}

type Holder struct {
	Err error
	Msg string
}

func zero() {}
func regular(a int, e error, c CustomErr) {}
func join(errs ...error) {}
func pair() (int, ValErr) { return 0, ValErr{} }
func errPair() (ValErr, *PtrErr) { return ValErr{}, &PtrErr{} }
func consume(a int, e error) {}
func box[T error](e T) {}

type errBox[E error] struct{ e E }

func (errBox[E]) m() {}

type errAlias[E error] = struct{ e E }

func genPair[A any, E error](_ A, _ E) {}

func compErrFunc[E interface {
	comparable
	error
}](_ E) {}

type customValErr struct{ ValErr }

func (customValErr) Code() int { return 0 }

func customErrFunc[E CustomErr](_ E) {}

func anyFunc[T any](_ T)         {}
func compFunc[T comparable](_ T) {}
func concreteFunc[T ValErr](_ T) {}
`

func parseAndCheck(t *testing.T, decls, body string) (*token.FileSet, string, *types.Info, *ast.File) {
	t.Helper()

	return parseAndCheckSource(t, commonDecls+"\n"+decls+"\nfunc f() {\n"+body+"\n}\n")
}

func parseAndCheckSource(t *testing.T, src string) (*token.FileSet, string, *types.Info, *ast.File) {
	t.Helper()

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "p.go", src, 0)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	info := &types.Info{
		Types:     make(map[ast.Expr]types.TypeAndValue),
		Instances: make(map[*ast.Ident]types.Instance),
		Uses:      make(map[*ast.Ident]types.Object),
		Defs:      make(map[*ast.Ident]types.Object),
	}

	conf := types.Config{Importer: importer.Default()}
	if _, err := conf.Check("p", fset, []*ast.File{file}, info); err != nil {
		t.Fatalf("typecheck failed: %v", err)
	}

	return fset, src, info, file
}

func snippet(fset *token.FileSet, src string, expr ast.Expr) string {
	start := fset.Position(expr.Pos()).Offset
	end := fset.Position(expr.End()).Offset

	return src[start:end]
}

func TestCallArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		decls string
		body  string
		want  []wantConv
	}{
		{
			name: "zero args",
			body: `zero()`,
			want: nil,
		},
		{
			name: "explicit error conversion",
			body: `_ = error(ValErr{})`,
			want: []wantConv{{typ: "p.ValErr", expr: "ValErr{}"}},
		},
		{
			name: "explicit non-error conversion",
			body: `type myInt int; _ = myInt(1)`,
			want: nil,
		},
		{
			name: "regular call skips non-error params",
			body: `var c CustomErr; regular(1, &PtrErr{}, c)`,
			want: []wantConv{
				{typ: "*p.PtrErr", expr: "&PtrErr{}"},
				{typ: "p.CustomErr", expr: "c"},
			},
		},
		{
			name: "variadic call",
			body: `join(ValErr{}, &PtrErr{})`,
			want: []wantConv{
				{typ: "p.ValErr", expr: "ValErr{}"},
				{typ: "*p.PtrErr", expr: "&PtrErr{}"},
			},
		},
		{
			name: "variadic ellipsis slice skipped",
			body: `var errs []error; join(errs...)`,
			want: nil,
		},
		{
			name: "multivalued tuple call",
			body: `consume(pair())`,
			want: []wantConv{{typ: "p.ValErr", expr: "pair()"}},
		},
		{
			name: "variadic single-arg tuple call",
			body: `join(errPair())`,
			want: []wantConv{
				{typ: "p.ValErr", expr: "errPair()"},
				{typ: "*p.PtrErr", expr: "errPair()"},
			},
		},
		{
			name: "generic call with concrete type param skipped",
			body: `box(&PtrErr{})`,
			want: nil,
		},
		{
			name: "generic call with error type arg",
			body: `box[error](&PtrErr{})`,
			want: []wantConv{{typ: "*p.PtrErr", expr: "&PtrErr{}"}},
		},
		{
			name:  "type parameter argument",
			decls: `func fwd[E error](e E) { join(e) }`,
			want:  []wantConv{{typ: "E", expr: "e"}},
		},
		{
			name:  "callee with core function type",
			decls: `func call[F ~func(int, error)](f F) { f(1, ValErr{}) }`,
			want:  []wantConv{{typ: "p.ValErr", expr: "ValErr{}"}},
		},
		{
			name:  "callee with core function type via embedded constraint",
			decls: `type errFunc interface{ ~func(error) }; func call[F interface{ errFunc; func(error) }](f F) { f(ValErr{}) }`,
			want:  []wantConv{{typ: "p.ValErr", expr: "ValErr{}"}},
		},
		{
			name: "builtin append",
			body: `var errs []error; errs = append(errs, ValErr{}); _ = errs`,
			want: []wantConv{{typ: "p.ValErr", expr: "ValErr{}"}},
		},
		{
			name: "builtin append string spread skipped",
			body: `var b []byte; b = append(b, "x"...); _ = b`,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fset, src, info, file := parseAndCheck(t, tt.decls, tt.body)

			var got []wantConv

			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				for typ, expr := range errorconv.CallArgs(info, call) {
					got = append(got, wantConv{
						typ:  typ.String(),
						expr: snippet(fset, src, expr),
					})
				}

				return true
			})

			if !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompositeLitElts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		decls string
		body  string
		want  []wantConv
	}{
		{
			name: "empty slice",
			body: `_ = []error{}`,
			want: nil,
		},
		{
			name: "slice keyed and unkeyed",
			body: `_ = []error{ValErr{}, 2: &PtrErr{}}`,
			want: []wantConv{
				{typ: "p.ValErr", expr: "ValErr{}"},
				{typ: "*p.PtrErr", expr: "&PtrErr{}"},
			},
		},
		{
			name: "array literal",
			body: `_ = [2]error{ValErr{}}`,
			want: []wantConv{{typ: "p.ValErr", expr: "ValErr{}"}},
		},
		{
			name: "non-error slice",
			body: `_ = []int{1, 2}`,
			want: nil,
		},
		{
			name: "map key and value",
			body: `_ = map[error]CustomErr{ValErr{}: CustomErr(nil)}`,
			want: []wantConv{
				{typ: "p.ValErr", expr: "ValErr{}"},
				{typ: "p.CustomErr", expr: "CustomErr(nil)"},
			},
		},
		{
			name: "map value only",
			body: `_ = map[string]error{"a": &PtrErr{}}`,
			want: []wantConv{{typ: "*p.PtrErr", expr: "&PtrErr{}"}},
		},
		{
			name: "struct keyed",
			body: `_ = Holder{Err: ValErr{}, Msg: "ok"}`,
			want: []wantConv{{typ: "p.ValErr", expr: "ValErr{}"}},
		},
		{
			name: "struct unkeyed",
			body: `_ = Holder{&PtrErr{}, "ok"}`,
			want: []wantConv{{typ: "*p.PtrErr", expr: "&PtrErr{}"}},
		},
		{
			name: "elided pointer to struct in slice",
			body: `_ = []*Holder{{Err: &PtrErr{}}}`,
			want: []wantConv{{typ: "*p.PtrErr", expr: "&PtrErr{}"}},
		},
		{
			name:  "slice of type parameter type",
			decls: `func lit[S ~[]error]() { _ = S{ValErr{}} }`,
			want:  []wantConv{{typ: "p.ValErr", expr: "ValErr{}"}},
		},
		{
			name:  "map of type parameter type",
			decls: `func lit[M ~map[string]error]() { _ = M{"a": &PtrErr{}} }`,
			want:  []wantConv{{typ: "*p.PtrErr", expr: "&PtrErr{}"}},
		},
		{
			name:  "struct of type parameter type",
			decls: `func lit[S ~struct{ Err error }]() { _ = S{Err: ValErr{}} }`,
			want:  []wantConv{{typ: "p.ValErr", expr: "ValErr{}"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fset, src, info, file := parseAndCheck(t, tt.decls, tt.body)

			var got []wantConv

			ast.Inspect(file, func(n ast.Node) bool {
				lit, ok := n.(*ast.CompositeLit)
				if !ok {
					return true
				}

				for typ, expr := range errorconv.CompositeLitElts(info, lit) {
					got = append(got, wantConv{
						typ:  typ.String(),
						expr: snippet(fset, src, expr),
					})
				}

				return true
			})

			if !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		decls string
		body  string
		want  []wantConv
	}{
		{
			name: "error channel",
			body: `var ch chan error; ch <- ValErr{}`,
			want: []wantConv{{typ: "p.ValErr", expr: "ValErr{}"}},
		},
		{
			name: "custom error channel",
			body: `var ch chan CustomErr; var c CustomErr; ch <- c`,
			want: []wantConv{{typ: "p.CustomErr", expr: "c"}},
		},
		{
			name: "non-error channel",
			body: `var ch chan int; ch <- 1`,
			want: nil,
		},
		{
			name:  "channel of type parameter type",
			decls: `func send[C ~chan error](ch C) { ch <- ValErr{} }`,
			want:  []wantConv{{typ: "p.ValErr", expr: "ValErr{}"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fset, src, info, file := parseAndCheck(t, tt.decls, tt.body)

			var got []wantConv

			ast.Inspect(file, func(n ast.Node) bool {
				send, ok := n.(*ast.SendStmt)
				if !ok {
					return true
				}

				if typ, expr, ok := errorconv.SendValue(info, send); ok {
					got = append(got, wantConv{
						typ:  typ.String(),
						expr: snippet(fset, src, expr),
					})
				}

				return true
			})

			if !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

type wantTypeArg struct {
	ident string
	obj   string
	typ   string
	index int
}

func TestTypeArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		decls string
		body  string
		want  []wantTypeArg
	}{
		{
			name: "generic defined type",
			body: `var _ errBox[*PtrErr]`,
			want: []wantTypeArg{{ident: "errBox", obj: "errBox", typ: "*p.PtrErr", index: 0}},
		},
		{
			name: "generic type alias",
			body: `var _ errAlias[ValErr]`,
			want: []wantTypeArg{{ident: "errAlias", obj: "errAlias", typ: "p.ValErr", index: 0}},
		},
		{
			name: "explicit generic function",
			body: `box[ValErr](ValErr{})`,
			want: []wantTypeArg{{ident: "box", obj: "box", typ: "p.ValErr", index: 0}},
		},
		{
			name: "inferred generic function",
			body: `box(&PtrErr{})`,
			want: []wantTypeArg{{ident: "box", obj: "box", typ: "*p.PtrErr", index: 0}},
		},
		{
			name: "multi-param mixed constraints",
			body: `genPair[int, *PtrErr](1, &PtrErr{})`,
			want: []wantTypeArg{{ident: "genPair", obj: "genPair", typ: "*p.PtrErr", index: 1}},
		},
		{
			name: "partial inference",
			body: `genPair[int](1, &PtrErr{})`,
			want: []wantTypeArg{{ident: "genPair", obj: "genPair", typ: "*p.PtrErr", index: 1}},
		},
		{
			name: "non-basic error constraint",
			body: `compErrFunc(ValErr{})`,
			want: []wantTypeArg{{ident: "compErrFunc", obj: "compErrFunc", typ: "p.ValErr", index: 0}},
		},
		{
			name: "custom error interface constraint",
			body: `customErrFunc(customValErr{})`,
			want: []wantTypeArg{{ident: "customErrFunc", obj: "customErrFunc", typ: "p.customValErr", index: 0}},
		},
		{
			name: "skips non-error constraints",
			body: `anyFunc(&PtrErr{}); compFunc(ValErr{}); concreteFunc(ValErr{})`,
			want: nil,
		},
		{
			name: "skips interface type arguments",
			body: `box[error](&PtrErr{}); var c CustomErr; customErrFunc(c)`,
			want: nil,
		},
		{
			name:  "forwarding call only yields outer instantiation",
			decls: `func fwd[E error](e E) { box[E](e) }`,
			body:  `fwd(ValErr{})`,
			want:  []wantTypeArg{{ident: "fwd", obj: "fwd", typ: "p.ValErr", index: 0}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			src := commonDecls + "\n" + tt.decls + "\nfunc f() {\n" + tt.body + "\n}\n"
			_, _, info, _ := parseAndCheckSource(t, src)

			var got []errorconv.TypeArg
			for arg := range errorconv.TypeArgs(info) {
				got = append(got, arg)
			}

			slices.SortFunc(got, func(a, b errorconv.TypeArg) int {
				return cmp.Or(cmp.Compare(a.Ident.Pos(), b.Ident.Pos()), cmp.Compare(a.Index, b.Index))
			})

			gotSummary := make([]wantTypeArg, len(got))
			for i, arg := range got {
				gotSummary[i] = wantTypeArg{
					ident: arg.Ident.Name,
					obj:   arg.Obj.Name(),
					typ:   arg.Type.String(),
					index: arg.Index,
				}
			}

			if !slices.Equal(gotSummary, tt.want) {
				t.Errorf("got %v, want %v", gotSummary, tt.want)
			}
		})
	}
}
