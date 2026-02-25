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

// TypeNameOf extracts the underlying type name from a given type.
// It handles pointers, dereferencing at most one pointer level to find the core [types.TypeName].
//
// It returns the found [types.TypeName], a boolean indicating if the original type
// was a pointer, and a boolean indicating if a type name was successfully found.
// It returns false for anonymous types (like struct literals), double pointers,
// type parameters, and basic types.
func TypeNameOf(t types.Type) (tn *types.TypeName, ptr, ok bool) {
	if p, isPtr := t.(*types.Pointer); isPtr {
		t, ptr = p.Elem(), true
	}

	switch typ := t.(type) {
	case *types.Named:
		return typ.Obj(), ptr, true

	case *types.Alias:
		return typ.Obj(), ptr, true

	default:
		// Anonymous types (struct literals, interface, etc.), double pointers,
		// type parameters, or basic types.
		return nil, ptr, false
	}
}

// CanonicalTypeNameOf is like [TypeNameOf], but resolves aliases (including
// chained and pointer aliases) as far as possible toward a non-generic defined
// type or a pointer to it: given
//
//	type (
//		t   struct{ _ int }
//		ta  = t
//		tp  = *t
//		tap = *ta
//		tpa = tp
//	)
//
//	func (t) f()     {}
//	func (*t) p()    {}
//	func (ta) fa()   {}
//	func (*ta) pa()  {}
//	func (tp) fp()   {}
//	func (tap) fap() {}
//	func (tpa) fpa() {}
//
// all receiver types resolve to t, with ptr set for pointer and pointer-alias
// types (*t, *ta, tp, tap, tpa).
//
// Resolution stops before stepping into a generic instantiation, an unnamed
// type, or a second pointer indirection, returning the last resolved alias.
// For example, given
//
//	type (
//		g[T any]  struct{ _ T }
//		gia       = g[int]
//		h         struct{ _ int }
//		hg[T any] = h
//		hi        = hg[int]
//		hia       = *hi
//	)
//
// gia resolves to gia, while both hi and hia resolve to hi (with ptr set for
// hia). By contrast, a defined "type hd hg[[int]]" is a new non-generic
// type, so "type hda = *hd" resolves to hd with ptr set. Note that Go does not
// allow [method declarations] on generic aliases or aliases of instantiated
// generic types (gia, hg[T], hi, *hi, hia).
//
// [method declarations]: https://go.dev/ref/spec#Method_declarations
func CanonicalTypeNameOf(t types.Type) (tn *types.TypeName, ptr, ok bool) {
	var (
		// speculative pointer flag for the next step
		rptr bool
		// whether tn holds a previously resolved alias
		aliased bool
	)

unwraploop:
	switch typ := t.(type) {
	case *types.Pointer:
		if rptr { // double pointer: stop resolving
			if aliased { // double pointer reached via alias (*PA or PPA = **T): keep last alias
				return tn, ptr, true
			}

			// literal double pointer (**T)
			return nil, true, false
		}

		// unwrap first pointer indirection
		t, rptr = typ.Elem(), true

		goto unwraploop

	case *types.Alias:
		if typ.TypeArgs().Len() > 0 { // generic alias
			if aliased { // generic alias instantiation reached via alias: keep last alias
				return tn, ptr, true
			}

			// top-level instantiated generic alias
			return typ.Obj(), rptr, true
		}

		// non-generic alias: record and step into RHS
		tn, ptr, aliased = typ.Obj(), rptr, true
		t = typ.Rhs()

		goto unwraploop

	case *types.Named:
		if aliased && typ.TypeArgs().Len() > 0 { // generic instantiation reached via alias: keep last alias
			return tn, ptr, true
		}

		// top-level generic or non-generic defined type
		return typ.Obj(), rptr, true

	default:
		if aliased { // unnamed type reached via alias: keep last alias
			return tn, ptr, true
		}

		// unnamed type, type parameter, or basic type
		return nil, rptr, false
	}
}
