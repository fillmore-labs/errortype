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
	"go/token"
	"go/types"
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
	// RecursiveStruct struct { RecursiveStruct }
	// RecursiveArray [1]RecursiveArray
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

	const (
		RecursiveStruct = "RecursiveStruct"
		RecursiveArray  = "RecursiveArray"
	)

	fset, f := parseSource(t, source)
	pkg, _ := checkSource(t, fset, []*ast.File{f})
	recursiveTypes := [...]types.Object{
		recursiveStructType(pkg, RecursiveStruct),
		recursiveArrayType(pkg, RecursiveArray),
	}

	scope := pkg.Scope()
	for _, obj := range recursiveTypes {
		scope.Insert(obj)
	}

	testCases := [...]struct {
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
		{RecursiveStruct, false},
		{RecursiveArray, false},
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
			if got := ZeroSized(typ); got != tc.isZero {
				t.Errorf("ZeroSized(%q) = %v; want %v", tc.name, got, tc.isZero)
			}
		})
	}
}

// recursiveStructType creates a recursively defined type within the given package and returns its type name.
// It defines a named struct type that contains a field of its own type, creating a nominally recursive type.
func recursiveStructType(pkg *types.Package, name string) *types.TypeName {
	typeName := types.NewTypeName(token.NoPos, pkg, name, nil)
	typ := types.NewNamed(typeName, nil, nil)

	field := types.NewField(token.NoPos, pkg, typeName.Name(), typ, true)
	underlying := types.NewStruct([]*types.Var{field}, nil)

	typ.SetUnderlying(underlying)

	return typeName
}

// recursiveArrayType creates a recursively defined type within the given package and returns its type name.
// It defines a named array type that contains an element of its own type, creating a nominally recursive type.
func recursiveArrayType(pkg *types.Package, name string) *types.TypeName {
	typeName := types.NewTypeName(token.NoPos, pkg, name, nil)
	typ := types.NewNamed(typeName, nil, nil)

	underlying := types.NewArray(typ, 1)

	typ.SetUnderlying(underlying)

	return typeName
}
