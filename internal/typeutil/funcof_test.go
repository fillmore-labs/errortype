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
	"slices"
	"testing"

	. "fillmore-labs.com/errortype/internal/typeutil"
)

func TestFuncOf(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []struct {
		name           string
		src            string
		wantFuncName   string
		wantMethodExpr bool
	}{
		{
			name: "simple function call",
			src: `func myFunc() int { return 0 }
var _ = myFunc()`,
			wantFuncName: "test.myFunc",
		},
		{
			name: "selector expression on package",
			src: `import "strings"
			var _ = strings.Clone("")`,
			wantFuncName: "strings.Clone",
		},
		{
			name: "method call on a variable",
			src: `type S struct{}
func (s S) myMethod() int { return 0 }
var v S
var _ = v.myMethod()`,
			wantFuncName: "(test.S).myMethod",
		},
		{
			name: "method expression call",
			src: `type S struct{}
func (s S) myMethod() int { return 0 }
var _ = (S).myMethod(S{})`,
			wantFuncName:   "(test.S).myMethod",
			wantMethodExpr: true,
		},
		{
			name: "method field call",
			src: `type S struct{ f func() int }
var v S
var _ = v.f()`,
		},
		{
			name: "generic function call with one type parameter",
			src: `func myFunc[T any]() T { return *new(T) }
var _ = myFunc[int]()`,
			wantFuncName: "test.myFunc",
		},
		{
			name: "generic function call with multiple type parameters",
			src: `func myFunc[T, U any]() T { return *new(T) }
var _ = myFunc[int, string]()`,
			wantFuncName: "test.myFunc",
		},
		{
			name: "parenthesized function call",
			src: `func myFunc() int { return 0 }
var _ = (myFunc)()`,
			wantFuncName: "test.myFunc",
		},
		{
			name: "call on function variable",
			src: `var myFunc func() int
var _ = myFunc()`,
		},
		{
			name: "call on a function pointer",
			src: `var myFunc *func() int
var _ = (*myFunc)()`,
		},
		{
			name: "call on a type conversion",
			src: `type myFuncType func() int
var f myFuncType
var _ = myFuncType(f)()`,
		},
		{
			name: "call on a function reult",
			src: `func myFunc() func() int { return nil }
var _ = (myFunc)()()`,
		},
		{
			name: "IndexExpr with non-type index",
			src: `var a [1]func() int
var _ = a[0]()`,
		},
	}

	if err := HasGo(); err != nil {
		tests = slices.Delete(tests, 1, 2)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			info, _, _, f := parseSource(t, tt.src)
			callExpr := lastDeclCallExpr(f)

			fun, _, methodExpr, ok := FuncOf(info, callExpr.Fun)

			wantOk := tt.wantFuncName != ""
			if ok != wantOk {
				t.Errorf("FuncOf() ok = %v, want %v", ok, wantOk)
			}

			if !wantOk {
				return
			}

			if fun == nil {
				t.Fatal("FuncOf() fun is nil, but wantOk is true")
			}

			funcName := fun.FullName()
			if funcName != tt.wantFuncName {
				t.Errorf("FuncOf() fun.FullName() = %q, want %q", funcName, tt.wantFuncName)
			}

			if methodExpr != tt.wantMethodExpr {
				t.Errorf("FuncOf() methodExpr = %v, want %v", methodExpr, tt.wantMethodExpr)
			}
		})
	}
}

//nolint:forcetypeassert
func lastDeclCallExpr(f *ast.File) *ast.CallExpr {
	lastDecl := f.Decls[len(f.Decls)-1]
	genDecl := lastDecl.(*ast.GenDecl)
	valSpec := genDecl.Specs[0].(*ast.ValueSpec)
	callExpr := valSpec.Values[0].(*ast.CallExpr)

	return callExpr
}
