// Copyright 2026 Oliver Eikemeier. All Rights Reserved.
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

// CoreType returns the underlying type of typ. For a type parameter, it returns
// the underlying type shared by all type terms of its constraint, or nil if
// there is none.
//
// This is conservative: channels of different directions, unions with interface
// terms and intersections that narrow differing unions report no core type.
func CoreType(typ types.Type) types.Type {
	switch t := types.Unalias(typ).(type) {
	case *types.Named:
		return t.Underlying()

	case *types.TypeParam:
		core, _ := termsCore(t.Constraint())

		return core

	default: // all other types are their own underlying type
		return t
	}
}

// termsCore returns the underlying type shared by all type terms in the
// interface constraint, including those of embedded interfaces. It returns
// (nil, true) when constraint has no type terms and (nil, false) when the terms
// don't share an underlying type.
func termsCore(constraint types.Type) (types.Type, bool) {
	iface, ok := constraint.Underlying().(*types.Interface)
	if !ok {
		return nil, false
	}

	var core types.Type

	merge := func(u types.Type) bool {
		switch {
		case u == nil:
			return true

		case core == nil:
			core = u

			return true

		default:
			return types.Identical(core, u)
		}
	}

	for embedded := range iface.EmbeddedTypes() {
		switch embedded := types.Unalias(embedded).(type) {
		case *types.Union: // ~T or T₁ | T₂ | ...
			for term := range embedded.Terms() {
				if termType := term.Type(); types.IsInterface(termType) || !merge(termType.Underlying()) {
					return nil, false
				}
			}

		default:
			if !types.IsInterface(embedded) { // single non-tilde term T
				if !merge(embedded.Underlying()) {
					return nil, false
				}

				continue
			}

			if embeddedCore, ok := termsCore(embedded); !ok || !merge(embeddedCore) {
				return nil, false
			}
		}
	}

	return core, true
}
