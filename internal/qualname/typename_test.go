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

package qualname_test

import (
	"cmp"
	"go/token"
	"go/types"
	"testing"

	. "fillmore-labs.com/errortype/internal/qualname"
)

func TestType_String(t *testing.T) {
	t.Parallel()

	tests := [...]struct {
		name     string
		typeName Type
		want     string
	}{
		{
			name:     "with path and name",
			typeName: Type{Path: "example.com/pkg", Name: "MyType"},
			want:     "example.com/pkg.MyType",
		},
		{
			name:     "with name only",
			typeName: Type{Name: "MyType"},
			want:     "MyType",
		},
		{
			name:     "empty",
			typeName: Type{},
			want:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.typeName.String(); got != tt.want {
				t.Errorf("Type.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestType_Compare(t *testing.T) {
	t.Parallel()

	tests := [...]struct {
		name   string
		p1, p2 Type
		want   int // -1, 0, or 1
	}{
		{
			name: "equal",
			p1:   Type{Path: "a/b", Name: "C"},
			p2:   Type{Path: "a/b", Name: "C"},
			want: 0,
		},
		{
			name: "path less",
			p1:   Type{Path: "a/a", Name: "C"},
			p2:   Type{Path: "a/b", Name: "C"},
			want: -1,
		},
		{
			name: "path greater",
			p1:   Type{Path: "a/c", Name: "C"},
			p2:   Type{Path: "a/b", Name: "C"},
			want: 1,
		},
		{
			name: "name less",
			p1:   Type{Path: "a/b", Name: "B"},
			p2:   Type{Path: "a/b", Name: "C"},
			want: -1,
		},
		{
			name: "name greater",
			p1:   Type{Path: "a/b", Name: "D"},
			p2:   Type{Path: "a/b", Name: "C"},
			want: 1,
		},
		{
			name: "path vs no path",
			p1:   Type{Path: "", Name: "C"},
			p2:   Type{Path: "a/b", Name: "C"},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := cmp.Compare(tt.p1.Compare(tt.p2), 0); got != tt.want {
				t.Errorf("Type.Compare() sign = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewType(t *testing.T) {
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

	tests := [...]struct {
		name string
		tn   *types.TypeName
		want Type
	}{
		{
			name: "type with package",
			tn:   createTestTypeName("example.com/user/project/pkg", "MyError"),
			want: Type{Path: "example.com/user/project/pkg", Name: "MyError"},
		},
		{
			name: "type from universe scope",
			tn:   createTestTypeName("", "error"),
			want: Type{Path: "", Name: "error"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := NewType(tt.tn); got != tt.want {
				t.Errorf("NewType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestType_MarshalText(t *testing.T) {
	t.Parallel()

	tests := [...]struct {
		name     string
		typeName Type
		want     string
	}{
		{
			name:     "with path",
			typeName: Type{Path: "example.com/pkg/path", Name: "MyType"},
			want:     "example.com/pkg/path.MyType",
		},
		{
			name:     "without path",
			typeName: Type{Path: "", Name: "error"},
			want:     "error",
		},
		{
			name:     "built-in type",
			typeName: Type{Path: "types", Name: "TypeName"},
			want:     "types.TypeName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.typeName.MarshalText()
			if err != nil {
				t.Fatalf("MarshalText() error = %v", err)
			}

			if string(got) != tt.want {
				t.Errorf("MarshalText() got = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestType_UnmarshalText(t *testing.T) {
	t.Parallel()

	tests := [...]struct {
		name    string
		text    string
		want    Type
		wantErr bool
	}{
		{"with path", "example.com/pkg/path.MyType", Type{Path: "example.com/pkg/path", Name: "MyType"}, false},
		{"without path", "error", Type{Path: "", Name: "error"}, false},
		{"built-in type", "types.TypeName", Type{Path: "types", Name: "TypeName"}, false},
		{"empty", "types.", Type{Path: "types", Name: ""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var tn Type
			if err := tn.UnmarshalText([]byte(tt.text)); (err != nil) != tt.wantErr {
				t.Fatalf("UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tn != tt.want {
				t.Errorf("UnmarshalText() got = %v, want %v", tn, tt.want)
			}
		})
	}
}
