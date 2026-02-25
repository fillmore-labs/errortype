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

package rules_test

import (
	"go/constant"
	"go/token"
	"go/types"
	"testing"

	. "fillmore-labs.com/errortype/internal/naming/rules"
)

func TestSuggest(t *testing.T) {
	t.Parallel()

	pkg := types.NewPackage("example.com/p", "p")

	newVar := func(name string) types.Object {
		return types.NewVar(token.NoPos, pkg, name, types.Universe.Lookup("error").Type())
	}
	newType := func(name string) types.Object {
		return types.NewTypeName(token.NoPos, pkg, name, types.Typ[types.Int])
	}

	tests := [...]struct {
		name         string
		obj          types.Object
		want         string
		wantNumbered string // Numbered(2), ignored when want is empty
	}{
		// Variables: the whole name precedes the count.
		{"var e", newVar("e"), "err", "err2"},
		{"var xErr", newVar("xErr"), "errX", "errX2"},
		{"var XErr", newVar("XErr"), "ErrX", "ErrX2"},
		{"var error", newVar("error"), "err", "err2"},
		{"var Error", newVar("Error"), "Err", "Err2"},
		{"var failure", newVar("failure"), "errFailure", "errFailure2"},
		{"var äErr", newVar("äErr"), "errÄ", "errÄ2"},
		{"const", types.NewConst(token.NoPos, pkg, "cErr", types.Typ[types.Int], constant.MakeInt64(1)), "errC", "errC2"},

		// Types: the count is inserted before the "Error" suffix.
		{"type xErr", newType("xErr"), "xError", "x2Error"},
		{"type MyErr", newType("MyErr"), "MyError", "My2Error"},
		{"type Error8", newType("Error8"), "E8Error", "E82Error"},
		{"type err8", newType("err8"), "e8Error", "e82Error"},

		// Empty stems keep the object's visibility.
		{"type err", newType("err"), "eError", "e2Error"},
		{"type Err", newType("Err"), "Error", "E2Error"},

		// "errors" is trimmed as a whole, not as "error" + "s".
		{"type errorsFoo", newType("errorsFoo"), "fooError", "foo2Error"},

		// Unsupported object kinds yield the zero Suggestion.
		{"label", types.NewLabel(token.NoPos, pkg, "xErr"), "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			suggestion := Suggest(tt.obj)

			if got := suggestion.Name(); got != tt.want {
				t.Errorf("Suggest(%s).Name() = %q, want %q", tt.obj.Name(), got, tt.want)
			}

			if tt.want == "" {
				return
			}

			if got := suggestion.Numbered(2); got != tt.wantNumbered {
				t.Errorf("Suggest(%s).Numbered(2) = %q, want %q", tt.obj.Name(), got, tt.wantNumbered)
			}
		})
	}
}
