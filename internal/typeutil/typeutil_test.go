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
	"go/types"
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

func TestHasSigs(t *testing.T) {
	t.Parallel()

	const src = `
		func sigError() string { return "" }
		func sigIs(error) bool { return false }
		func sigAs(any) bool { return false }
		func sigUnwrap() error { return nil }
		func sigUnwrapMultiple() []error { return nil }

		func sigNoParamsNoResults() {}
		func sigWrongParam(int) bool { return false }
		func sigWrongResult() int { return 0 }
		func sigTooManyParams(error, error) bool { return false }
		func sigTooManyResults() (string, string) { return "", "" }
	`

	_, pkg, _, _ := parseSource(t, src)

	getSig := func(name string) *types.Signature {
		t.Helper()

		obj := pkg.Scope().Lookup(name)
		if obj == nil {
			t.Fatalf("function %q not found in test source", name)
		}

		fun, ok := obj.(*types.Func)
		if !ok {
			t.Fatalf("object %q is not a function", name)
		}

		return fun.Type().(*types.Signature)
	}

	tests := []struct {
		name    string
		sigName string
		checker func(*types.Signature) bool
		want    bool
	}{
		// HasErrorSig
		{"HasErrorSig: correct", "sigError", HasErrorSig, true},
		{"HasErrorSig: wrong result", "sigWrongResult", HasErrorSig, false},
		{"HasErrorSig: with params", "sigIs", HasErrorSig, false},

		// HasIsSig
		{"HasIsSig: correct", "sigIs", HasIsSig, true},
		{"HasIsSig: wrong param", "sigWrongParam", HasIsSig, false},
		{"HasIsSig: no params", "sigError", HasIsSig, false},
		{"HasIsSig: too many params", "sigTooManyParams", HasIsSig, false},

		// HasAsSig
		{"HasAsSig: correct", "sigAs", HasAsSig, true},
		{"HasAsSig: wrong param", "sigWrongParam", HasAsSig, false},
		{"HasAsSig: no params", "sigError", HasAsSig, false},

		// HasUnwrapSig
		{"HasUnwrapSig: correct", "sigUnwrap", HasUnwrapSig, true},
		{"HasUnwrapSig: multiple correct", "sigUnwrapMultiple", HasUnwrapSig, true},
		{"HasUnwrapSig: wrong result", "sigWrongResult", HasUnwrapSig, false},
		{"HasUnwrapSig: with params", "sigIs", HasUnwrapSig, false},

		// General negative cases
		{"General: no params, no results", "sigNoParamsNoResults", HasErrorSig, false},
		{"General: too many results", "sigTooManyResults", HasErrorSig, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sig := getSig(tt.sigName)
			if got := tt.checker(sig); got != tt.want {
				t.Errorf("check failed: got %v, want %v", got, tt.want)
			}
		})
	}
}
