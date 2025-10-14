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
	"go/types"
	"testing"

	. "fillmore-labs.com/errortype/internal/typeutil"
)

func TestComparable(t *testing.T) {
	t.Parallel()

	tests := [...]struct {
		name string
		typ  types.Type
		want bool
	}{
		{
			name: "basic int",
			typ:  types.Typ[types.Int],
			want: true,
		},
		{
			name: "slice of int",
			typ:  types.NewSlice(types.Typ[types.Int]),
			want: false,
		},
		{
			name: "map int->string",
			typ:  types.NewMap(types.Typ[types.Int], types.Typ[types.String]),
			want: false,
		},
		{
			name: "simple struct with int field",
			typ: types.NewStruct([]*types.Var{
				types.NewField(0, nil, "A", types.Typ[types.Int], false),
			}, nil),
			want: true,
		},
		{
			name: "struct with slice field",
			typ: types.NewStruct([]*types.Var{
				types.NewField(0, nil, "A", types.NewSlice(types.Typ[types.Int]), false),
			}, nil),
			want: false,
		},
		{
			name: "array of int",
			typ:  types.NewArray(types.Typ[types.Int], 5),
			want: true,
		},
		{
			name: "array of slices",
			typ:  types.NewArray(types.NewSlice(types.Typ[types.Int]), 5),
			want: false,
		},
		{
			name: "interface type",
			typ:  types.NewInterfaceType(nil, nil).Complete(),
			want: true,
		},
		{
			name: "signature function type",
			typ:  types.NewSignatureType(nil, nil, nil, nil, nil, false),
			want: false,
		},
		{
			name: "tuple type",
			typ:  types.NewTuple(types.NewParam(0, nil, "", types.Typ[types.Int])),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := Comparable(tt.typ); got != tt.want {
				t.Errorf("Comparable(%v) = %v, want %v", tt.typ, got, tt.want)
			}
		})
	}
}
