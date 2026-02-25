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

	fset, f := parseSource(t, src)
	pkg, _ := checkSource(t, fset, []*ast.File{f})

	scope := pkg.Scope()
	sigOf := func(name string) *types.Signature {
		t.Helper()

		obj := scope.Lookup(name)
		if obj == nil {
			t.Fatalf("function %q not found in test source", name)
		}

		fun, ok := obj.(*types.Func)
		if !ok {
			t.Fatalf("object %q is not a function", name)
		}

		return fun.Signature()
	}

	tests := [...]struct {
		name    string
		method  Method
		sigName string
		want    bool
	}{
		// HasErrorSig
		{"HasErrorSig: correct", ErrorMethod, "sigError", true},
		{"HasErrorSig: wrong result", ErrorMethod, "sigWrongResult", false},
		{"HasErrorSig: with params", ErrorMethod, "sigIs", false},

		// HasIsSig
		{"HasIsSig: correct", IsMethod, "sigIs", true},
		{"HasIsSig: wrong param", IsMethod, "sigWrongParam", false},
		{"HasIsSig: no params", IsMethod, "sigError", false},
		{"HasIsSig: too many params", IsMethod, "sigTooManyParams", false},

		// HasAsSig
		{"HasAsSig: correct", AsMethod, "sigAs", true},
		{"HasAsSig: wrong param", AsMethod, "sigWrongParam", false},
		{"HasAsSig: no params", AsMethod, "sigError", false},

		// HasUnwrapSig
		{"HasUnwrapSig: correct", UnwrapMethod, "sigUnwrap", true},
		{"HasUnwrapSig: multiple correct", UnwrapMethod, "sigUnwrapMultiple", true},
		{"HasUnwrapSig: wrong result", UnwrapMethod, "sigWrongResult", false},
		{"HasUnwrapSig: with params", UnwrapMethod, "sigIs", false},

		// HasUnwrapMul
		{"HasUnwrapMultipleSig: correct", UnwrapMultipleMethod, "sigUnwrapMultiple", true},
		{"HasUnwrapMultipleSig: single", UnwrapMultipleMethod, "sigUnwrap", false},

		// General negative cases
		{"General: no params, no results", ErrorMethod, "sigNoParamsNoResults", false},
		{"General: too many results", ErrorMethod, "sigTooManyResults", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sig := sigOf(tt.sigName)
			if got := tt.method.MatchSig(sig); got != tt.want {
				t.Errorf("check failed: got %v, want %v", got, tt.want)
			}
		})
	}
}
