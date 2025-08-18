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
	"testing"

	. "fillmore-labs.com/errortype/internal/typeutil"
)

func TestZeroSized(t *testing.T) {
	t.Parallel()

	source := `
type (
	EmptyStruct               struct{}
	ZeroArray                 [0]int
	NonZeroArray              [1]int
	StructWithZeroSizedFields struct {
		f1 struct{}
		f2 [0]string
	}
	StructWithNonZeroSizedField struct {
		f1 struct{}
		f2 int
	}
	NestedZeroSized struct {
		f1 StructWithZeroSizedFields
	}
)

var (
	emptyStruct struct{}
	basic       int
	ptr         *int
	slice       []int
	amap        map[string]int
	achan       chan int
	aninterface interface{}
	afunc       func()
)
`
	_, pkg, _, _ := parseSource(t, source)
	scope := pkg.Scope()

	testCases := []struct {
		name   string
		isZero bool
	}{
		{"emptyStruct", true},
		{"EmptyStruct", true},
		{"ZeroArray", true},
		{"NonZeroArray", false},
		{"StructWithZeroSizedFields", true},
		{"StructWithNonZeroSizedField", false},
		{"NestedZeroSized", true},
		{"basic", false},
		{"ptr", false},
		{"slice", false},
		{"amap", false},
		{"achan", false},
		{"aninterface", false},
		{"afunc", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			obj := scope.Lookup(tc.name)
			if obj == nil {
				t.Fatalf("type %q not found", tc.name)
			}

			typ := obj.Type()
			if got := ZeroSized(typ, 0); got != tc.isZero {
				t.Errorf("ZeroSized(%q) = %v; want %v", tc.name, got, tc.isZero)
			}
		})
	}
}
