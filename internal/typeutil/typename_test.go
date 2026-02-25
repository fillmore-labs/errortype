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
	"cmp"
	"go/ast"
	"go/token"
	"go/types"
	"testing"

	. "fillmore-labs.com/errortype/internal/typeutil"
)

func TestTypeName_String(t *testing.T) {
	t.Parallel()

	tests := [...]struct {
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

	tests := [...]struct {
		name   string
		p1, p2 TypeName
		want   int // -1, 0, or 1
	}{
		{
			name: "equal",
			p1:   TypeName{Path: "a/b", Name: "C"},
			p2:   TypeName{Path: "a/b", Name: "C"},
			want: 0,
		},
		{
			name: "path less",
			p1:   TypeName{Path: "a/a", Name: "C"},
			p2:   TypeName{Path: "a/b", Name: "C"},
			want: -1,
		},
		{
			name: "path greater",
			p1:   TypeName{Path: "a/c", Name: "C"},
			p2:   TypeName{Path: "a/b", Name: "C"},
			want: 1,
		},
		{
			name: "name less",
			p1:   TypeName{Path: "a/b", Name: "B"},
			p2:   TypeName{Path: "a/b", Name: "C"},
			want: -1,
		},
		{
			name: "name greater",
			p1:   TypeName{Path: "a/b", Name: "D"},
			p2:   TypeName{Path: "a/b", Name: "C"},
			want: 1,
		},
		{
			name: "path vs no path",
			p1:   TypeName{Path: "", Name: "C"},
			p2:   TypeName{Path: "a/b", Name: "C"},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := cmp.Compare(tt.p1.Compare(tt.p2), 0); got != tt.want {
				t.Errorf("TypeName.Compare() sign = %v, want %v", got, tt.want)
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

	tests := [...]struct {
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

func TestTypeName_MarshalText(t *testing.T) {
	t.Parallel()

	tests := [...]struct {
		name     string
		typeName TypeName
		want     string
	}{
		{
			name:     "with path",
			typeName: TypeName{Path: "example.com/pkg/path", Name: "MyType"},
			want:     "example.com/pkg/path.MyType",
		},
		{
			name:     "without path",
			typeName: TypeName{Path: "", Name: "error"},
			want:     "error",
		},
		{
			name:     "built-in type",
			typeName: TypeName{Path: "types", Name: "TypeName"},
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

func TestTypeName_UnmarshalText(t *testing.T) {
	t.Parallel()

	tests := [...]struct {
		name    string
		text    string
		want    TypeName
		wantErr bool
	}{
		{"with path", "example.com/pkg/path.MyType", TypeName{Path: "example.com/pkg/path", Name: "MyType"}, false},
		{"without path", "error", TypeName{Path: "", Name: "error"}, false},
		{"built-in type", "types.TypeName", TypeName{Path: "types", Name: "TypeName"}, false},
		{"empty", "types.", TypeName{Path: "types", Name: ""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var tn TypeName
			if err := tn.UnmarshalText([]byte(tt.text)); (err != nil) != tt.wantErr {
				t.Fatalf("UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tn != tt.want {
				t.Errorf("UnmarshalText() got = %v, want %v", tn, tt.want)
			}
		})
	}
}

func TestTypeNameOf(t *testing.T) {
	t.Parallel()

	source := `
type (
	T struct{}
	A = T
)

var (
	vT       T
	vPtrT    *T
	vA       A
	vPtrPtrT **T
	vStruct  struct{}
)
`

	fset, f := parseSource(t, source)
	pkg, _ := checkSource(t, fset, []*ast.File{f})

	tests := [...]struct {
		variable string
		wantName string
		wantPtr  bool
		wantOK   bool
	}{
		{"vT", "T", false, true},
		{"vPtrT", "T", true, true},
		{"vA", "A", false, true},
		{"vPtrPtrT", "", true, false},
		{"vStruct", "", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.variable, func(t *testing.T) {
			t.Parallel()

			obj := pkg.Scope().Lookup(tt.variable)
			if obj == nil {
				t.Fatalf("variable %q not found", tt.variable)
			}

			tn, ptr, ok := TypeNameOf(obj.Type())
			if ok != tt.wantOK {
				t.Fatalf("TypeNameOf(%s) ok = %t, want %t", obj.Type(), ok, tt.wantOK)
			}

			if ptr != tt.wantPtr {
				t.Errorf("TypeNameOf(%s) ptr = %t, want %t", obj.Type(), ptr, tt.wantPtr)
			}

			var gotName string
			if tn != nil {
				gotName = tn.Name()
			}

			if gotName != tt.wantName {
				t.Errorf("TypeNameOf(%s) name = %q, want %q", obj.Type(), gotName, tt.wantName)
			}
		})
	}
}

func TestCanonicalTypeNameOf(t *testing.T) {
	t.Parallel()

	source := `
type (
	T        struct{}
	G[P any] struct{}

	A   = T
	AA  = A
	PA  = *T
	PAA = *A
	IA  = G[int]
	UA  = struct{ T }
	PUA = *struct{ T }

	GA[P any] = G[P]

	HG[P any] = T
	HI        = HG[int]
	HIA       = *HI
	HD        HG[int]
	HDA       = *HD
)

var (
	vT     T
	vA     A
	vAA    AA
	vPtrA  *A
	vPA    PA
	vPAA   PAA
	vPtrPA *PA
	vIA    IA
	vPtrIA *IA
	vUA    UA
	vPUA   PUA
	vGA    GA[int]
	vG     G[int]
	vHI    HI
	vHIA   HIA
	vHD    HD
	vHDA   HDA
)
`

	fset, f := parseSource(t, source)
	pkg, _ := checkSource(t, fset, []*ast.File{f})

	tests := [...]struct {
		variable string
		wantName string
		wantPtr  bool
	}{
		{"vT", "T", false},
		{"vA", "T", false},
		{"vAA", "T", false},
		{"vPtrA", "T", true},
		{"vPA", "T", true},
		{"vPAA", "T", true},
		{"vPtrPA", "PA", true}, // double pointer, keep alias
		{"vIA", "IA", false},   // generic instantiation
		{"vPtrIA", "IA", true}, // generic instantiation
		{"vUA", "UA", false},   // unnamed type
		{"vPUA", "PUA", false}, // unnamed type
		{"vGA", "GA", false},   // generic alias
		{"vG", "G", false},
		{"vHI", "HI", false}, // type args in alias chain, keep alias
		{"vHIA", "HI", true}, // resolves to *HI, stops before HG[int]
		{"vHD", "HD", false}, // defined type over instantiated generic alias
		{"vHDA", "HD", true}, // pointer alias to defined type HD
	}
	for _, tt := range tests {
		t.Run(tt.variable, func(t *testing.T) {
			t.Parallel()

			obj := pkg.Scope().Lookup(tt.variable)
			if obj == nil {
				t.Fatalf("variable %q not found", tt.variable)
			}

			tn, ptr, ok := CanonicalTypeNameOf(obj.Type())
			if !ok {
				t.Fatalf("CanonicalTypeNameOf(%s) not ok", obj.Type())
			}

			if tn.Name() != tt.wantName || ptr != tt.wantPtr {
				t.Errorf("CanonicalTypeNameOf(%s) = %s, %t, want %s, %t", obj.Type(), tn.Name(), ptr, tt.wantName, tt.wantPtr)
			}
		})
	}
}
