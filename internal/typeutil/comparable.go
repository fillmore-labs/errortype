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

package typeutil

import (
	"go/types"
	"iter"
)

// Comparable determines if a given type is [comparable].
// Structs are checked recursively to ensure all fields are comparable.
// Simplistic assumption is made for interfaces being always comparable,
// not considering type parameters.
//
// [comparable]: https://go.dev/ref/spec#Comparison_operators
func Comparable(t types.Type) bool {
	for {
		switch u := t.Underlying().(type) {
		case *types.Slice, *types.Map, *types.Signature, *types.Tuple:
			return false

		case *types.Struct:
			return all(u.Fields(), isComparable)

		case *types.Array:
			t = u.Elem()

		default:
			return true
		}
	}
}

func isComparable[O types.Object](obj O) bool { return Comparable(obj.Type()) }

func all[T any](seq iter.Seq[T], predicate func(T) bool) bool {
	for v := range seq {
		if !predicate(v) {
			return false
		}
	}

	return true
}
