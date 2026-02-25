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
