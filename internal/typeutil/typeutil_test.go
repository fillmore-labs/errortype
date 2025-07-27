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
	"testing"

	. "fillmore-labs.com/errortype/internal/typeutil"
)

func TestHasErrorResult(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		src       string
		funcName  string
		wantIndex int
	}{
		{
			name:      "no return values",
			src:       `func noReturn() {}`,
			funcName:  "noReturn",
			wantIndex: -1,
		},
		{
			name:      "single error return",
			src:       `func singleError() error { return nil }`,
			funcName:  "singleError",
			wantIndex: 0,
		},
		{
			name:      "multiple returns, last is error",
			src:       `func multiReturnWithError() (int, error) { return 0, nil }`,
			funcName:  "multiReturnWithError",
			wantIndex: 1,
		},
		{
			name: "multiple returns, last is custom value error",
			src: `
type MyError struct{}
func (e MyError) Error() string { return "my error" }
func customError() (int, interface { error }) { return 0, MyError{} }`,
			funcName:  "customError",
			wantIndex: 1,
		},
		{
			name:      "single return, not error",
			src:       `func singleNonError() int { return 0 }`,
			funcName:  "singleNonError",
			wantIndex: -1,
		},
		{
			name:      "multiple returns, last is not error",
			src:       `func multiReturnNotError() (error, int) { return nil, 0 }`,
			funcName:  "multiReturnNotError",
			wantIndex: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			info, _, _, f := parseSource(t, tt.src)
			funcDecl := findFunc(t, f, tt.funcName)

			index := HasErrorResult(info, funcDecl.Type.Results)

			if index != tt.wantIndex {
				t.Errorf("HasErrorResult() index = %v, want %v", index, tt.wantIndex)
			}
		})
	}
}

func findFunc(tb testing.TB, f *ast.File, name string) *ast.FuncDecl {
	tb.Helper()

	for _, decl := range f.Decls {
		d, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if d.Name.Name == name {
			return d
		}
	}

	tb.Fatalf("function %q not found in test source", name)

	return nil
}
