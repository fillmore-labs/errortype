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
	"go/token"
	"go/types"
	"testing"

	. "fillmore-labs.com/errortype/internal/typeutil"
)

func TestTypeName_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		typeName TypeName
		want     string
	}{
		{
			name:     "with path and name",
			typeName: TypeName{Path: "example.com/pkg", Name: "MyType"},
			want:     "example.com/pkg.MyType",
		},
		{
			name:     "with name only",
			typeName: TypeName{Name: "MyType"},
			want:     "MyType",
		},
		{
			name:     "empty",
			typeName: TypeName{},
			want:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.typeName.String(); got != tt.want {
				t.Errorf("TypeName.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTypeName_Compare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		p1    TypeName
		p2    TypeName
		want  int // -1, 0, or 1
		check func(int) bool
	}{
		{
			name:  "equal",
			p1:    TypeName{Path: "a/b", Name: "C"},
			p2:    TypeName{Path: "a/b", Name: "C"},
			check: func(i int) bool { return i == 0 },
		},
		{
			name:  "path less",
			p1:    TypeName{Path: "a/a", Name: "C"},
			p2:    TypeName{Path: "a/b", Name: "C"},
			check: func(i int) bool { return i < 0 },
		},
		{
			name:  "path greater",
			p1:    TypeName{Path: "a/c", Name: "C"},
			p2:    TypeName{Path: "a/b", Name: "C"},
			check: func(i int) bool { return i > 0 },
		},
		{
			name:  "name less",
			p1:    TypeName{Path: "a/b", Name: "B"},
			p2:    TypeName{Path: "a/b", Name: "C"},
			check: func(i int) bool { return i < 0 },
		},
		{
			name:  "name greater",
			p1:    TypeName{Path: "a/b", Name: "D"},
			p2:    TypeName{Path: "a/b", Name: "C"},
			check: func(i int) bool { return i > 0 },
		},
		{
			name:  "path vs no path",
			p1:    TypeName{Path: "", Name: "C"},
			p2:    TypeName{Path: "a/b", Name: "C"},
			check: func(i int) bool { return i < 0 },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.p1.Compare(tt.p2); !tt.check(got) {
				t.Errorf("TypeName.Compare() = %v, did not meet check", got)
			}
		})
	}
}

func TestNewTypeName(t *testing.T) {
	t.Parallel()

	// Helper to create a types.TypeName for testing
	createTestTypeName := func(pkgPath, typeName string) *types.TypeName {
		var pkg *types.Package
		if pkgPath != "" {
			pkg = types.NewPackage(pkgPath, "main")
		}
		// The underlying type doesn't matter for this test
		return types.NewTypeName(token.NoPos, pkg, typeName, nil)
	}

	tests := []struct {
		name string
		tn   *types.TypeName
		want TypeName
	}{
		{
			name: "type with package",
			tn:   createTestTypeName("example.com/user/project/pkg", "MyError"),
			want: TypeName{Path: "example.com/user/project/pkg", Name: "MyError"},
		},
		{
			name: "type from universe scope",
			tn:   createTestTypeName("", "error"),
			want: TypeName{Path: "", Name: "error"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := NewTypeName(tt.tn); got != tt.want {
				t.Errorf("NewTypeName() = %v, want %v", got, tt.want)
			}
		})
	}
}
