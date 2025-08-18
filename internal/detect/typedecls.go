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

package detect

import (
	"go/types"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// processTypeDecls analyzes all type declarations in the current package, identifying types
// that implement the error interface either directly or via embedding.
//
// For each such type, it determines whether the "Error" method has a pointer or value receiver,
// and records this property in the propertyMap. Types that are interfaces are skipped.
func (p pass) processTypeDecls() {
	for typespec := range p.AllTypeDecls {
		tn, ok := p.TypesInfo.Defs[typespec.Name].(*types.TypeName)
		if !ok { // should not happen
			p.LogErrorf(typespec.Name, "Not a TypeName: %s", typespec.Name.Name)

			continue
		}

		obj, _, indirect := types.LookupFieldOrMethod(tn.Type(), true, p.Pkg, "Error")
		if obj == nil {
			continue // No "Error" method
		}

		fun, ok := obj.(*types.Func)
		if !ok || !typeutil.HasErrorSig(fun.Signature()) {
			continue // *types.Var or wrong signature, not an error type
		}

		prop := None

		switch u := tn.Type().Underlying().(type) {
		case *types.Interface:
			continue // Interface type

		case *types.Struct:
			if typeutil.ZeroSized(u, 0) {
				prop |= ZeroSized
			}

		case *types.Pointer:
			// The type is an alias of a pointer to type with an `Error() string` method.
			// This should be rare.
			prop |= PointerDef
			if typeutil.ZeroSized(u.Elem(), 0) {
				prop |= ZeroSized
			}

		default:
			// Non-Struct error types are often value types
			prop |= NonStruct
		}

		_, ptrRecv := typeutil.HasPointerReceiver(fun.Signature())

		if ptrRecv && !indirect {
			// Type has a `Error() string` method with a pointer receiver, possibly embedded without indirections
			prop |= PointerReceiver
		}

		// Otherwise the type has a (possibly embedded) `Error() string` method, either with value receiver
		// or the receiver type is not relevant because of indirection. We need to rely on heuristics.
		p.AddTypeProperty(tn, prop)
	}
}
