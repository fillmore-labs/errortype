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
	"go/version"
	"runtime"
	"testing"

	. "fillmore-labs.com/errortype/internal/typeutil"
)

func TestFuncOf(t *testing.T) {
	t.Parallel()

	tests := [...]struct {
		name           string
		src            string
		wantFuncName   string
		version        string
		wantMethodExpr bool
	}{
		{
			name:         "simple function call",
			src:          `func myFunc() int { return 0 }; var _ = myFunc()`,
			wantFuncName: "test.myFunc",
		},
		{
			name:         "selector expression on package",
			src:          `import "strings"; var _ = strings.Clone("")`,
			wantFuncName: "strings.Clone",
		},
		{
			name:         "method call on a variable",
			src:          `type S struct{}; func (s S) myMethod() int { return 0 }; var v S; var _ = v.myMethod()`,
			wantFuncName: "(test.S).myMethod",
		},
		{
			name:         "method call on a interface variable",
			src:          `type S interface{ myMethod() int }; var v S; var _ = v.myMethod()`,
			wantFuncName: "(test.S).myMethod",
		},
		{
			name:           "method expression call",
			src:            `type S struct{}; func (s S) myMethod() int { return 0 }; var v S; var _ = (S).myMethod(v)`,
			wantFuncName:   "(test.S).myMethod",
			wantMethodExpr: true,
		},
		{
			name:           "method expression call on generic type",
			src:            `type S[T any] struct{}; func (s S[T]) myMethod() int { return 0 }; var v S[int]; var _ = (S[int]).myMethod(v)`,
			wantFuncName:   "(test.S).myMethod",
			wantMethodExpr: true,
		},
		{
			name:         "embendded method call",
			src:          `type S struct{}; func (s S) myMethod() int { return 0 }; type T struct{ S }; var v T; var _ = v.myMethod()`,
			wantFuncName: "(test.S).myMethod",
		},
		{
			name:           "embendded method expression call",
			src:            `type S struct{}; func (s S) myMethod() int { return 0 }; type T struct{ S }; var v T; var _ = (T).myMethod(v)`,
			wantFuncName:   "(test.S).myMethod",
			wantMethodExpr: true,
		},
		{
			name: "method field call",
			src:  `type S struct{ f func() int }; var v S; var _ = v.f()`,
		},
		{
			name:         "generic function call with one type parameter",
			src:          `func myFunc[T any]() T { var t T; return t }; var _ = myFunc[int]()`,
			wantFuncName: "test.myFunc",
		},
		{
			name:         "generic function call with multiple type parameters",
			src:          `func myFunc[T, U any]() T { var t T; return t }; var _ = myFunc[int, string]()`,
			wantFuncName: "test.myFunc",
		},
		{
			name:           "generic method call",
			src:            `type S[T any] struct{}; func (s S[T]) myMethod[U any]() int { return 0 }; var v S[int]; var _ = (S[int]).myMethod[string](v)`,
			wantFuncName:   "(test.S).myMethod",
			wantMethodExpr: true,
			version:        "go1.27",
		},
		{
			name:           "generic non-instantiated method call",
			src:            `type S[T any] struct{}; func (s S[T]) myMethod[U any]() int { return 0 };  func _[T, U any]() { var v S[T]; _ = (S[T]).myMethod[U](v) }`,
			wantFuncName:   "(test.S).myMethod",
			wantMethodExpr: true,
			version:        "go1.27",
		},
		{
			name:         "parenthesized function call",
			src:          `func myFunc() int { return 0 }; var _ = (myFunc)()`,
			wantFuncName: "test.myFunc",
		},
		{
			name: "builtin new",
			src:  `var _ = new(int)`,
		},
		{
			name:    "builtin new expr",
			src:     `var _ = new(1)`,
			version: "go1.26",
		},
		{
			name:         "call on function variable",
			src:          `var myFunc func() int; var _ = myFunc()`,
			wantFuncName: "test.myFunc",
		},
		{
			name: "call on a function pointer",
			src:  `var myFunc *func() int; var _ = (*myFunc)()`,
		},
		{
			name: "type conversion",
			src:  `type myFuncType func() int; var f myFuncType; var _ = myFuncType(f)`,
		},
		{
			name: "external type conversion",
			src:  `import "go/doc"; var _ = doc.Filter(nil)`,
		},
		{
			name: "call on a function result",
			src:  `func myFunc() func() int { return nil }; var _ = (myFunc)()()`,
		},
		{
			name: "IndexExpr with non-type index",
			src:  `var a [1]func() int; var _ = a[0]()`,
		},
	}

	current := version.Lang(runtime.Version())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.version != "" {
				if version.Compare(tt.version, current) > 0 {
					t.Skipf("test needs at least Go %s", tt.version)
				}
			}

			fset, f := parseSource(t, tt.src)
			_, info := checkSource(t, fset, []*ast.File{f})
			call := lastDeclCallExpr(f)

			fun, ok := FuncOf(info, call)

			wantOk := tt.wantFuncName != ""
			if ok != wantOk {
				t.Errorf("FuncOf() ok = %v, want %v", ok, wantOk)
			}

			if !wantOk {
				return
			}

			if fun.Func == nil {
				t.Fatal("FuncOf() fun is nil, but wantOk is true")
			}

			if funcName := FuncNameOf(fun.Func).String(); funcName != tt.wantFuncName {
				t.Errorf("FuncOf() FuncNameOf = %q, want %q", funcName, tt.wantFuncName)
			}

			if fun.MethodExpr != tt.wantMethodExpr {
				t.Errorf("FuncOf() methodExpr = %v, want %v", fun.MethodExpr, tt.wantMethodExpr)
			}
		})
	}
}

func lastDeclCallExpr(file *ast.File) *ast.CallExpr {
	for node := range ast.Preorder(file) {
		if call, ok := node.(*ast.CallExpr); ok {
			return call
		}
	}

	return nil
}
