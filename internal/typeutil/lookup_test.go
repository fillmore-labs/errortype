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

	"fillmore-labs.com/errortype/internal/typeutil"
)

func TestHasErrorMethod(t *testing.T) {
	t.Parallel()

	const src = `
		type A struct{}
		func (A) Error() string { return "" }

		type B struct{}
		func (*B) Error() string { return "" }

		type C interface { Error() string }
		type D interface { comparable; Error() string }
	`

	fset, f := parseSource(t, src)
	pkg, _ := checkSource(t, fset, []*ast.File{f})

	scope := pkg.Scope()

	tests := [...]struct {
		name string
		want bool
	}{
		{"A", true},
		{"B", false},
		{"C", true},
		{"D", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			obj := scope.Lookup(tt.name)
			if obj == nil {
				t.Fatalf("type %q not found in test source", tt.name)
			}

			typ := obj.Type()
			if got, want := typeutil.HasErrorMethod(typ), tt.want; got != want {
				t.Errorf("HasErrorMethod failed: got %t, want %t", got, want)
			}
		})
	}
}
