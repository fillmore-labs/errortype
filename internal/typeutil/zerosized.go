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

import "go/types"

const maxDepth = 10

// ZeroSized determines whether the type t is provably zero-sized.
func ZeroSized(typ types.Type, depth int) bool {
	if depth > maxDepth {
		return false
	}

	switch u := typ.Underlying().(type) {
	case *types.Array:
		if u.Len() > 0 {
			return ZeroSized(u.Elem(), depth+1)
		}

		return true

	case *types.Struct:
		for i := range u.NumFields() {
			if !ZeroSized(u.Field(i).Type(), depth+1) {
				return false
			}
		}

		return true

	default:
		return false
	}
}
