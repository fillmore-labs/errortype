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
	"context"
	"go/types"
	"runtime/trace"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// processTypeDecls analyzes all type declarations in the current package, identifying types
// that implement the error interface either directly or via embedding.
//
// For each such type, it determines whether the "Error" method has a pointer or value receiver,
// and records this property in the propertyMap. Types that are interfaces are skipped.
func (p pass) processTypeDecls(ctx context.Context) {
	defer trace.StartRegion(ctx, "typeDecls").End()

	for typeDecl := range p.AllTypeDecls {
		tn, ok := p.TypesInfo.Defs[typeDecl.Name].(*types.TypeName)
		if !ok { // should not happen
			p.LogErrorf(typeDecl.Name, "Not a TypeName: %s", typeDecl.Name.Name)

			continue
		}

		fun, indirect, embedded, ok := p.LookupMethod(tn, "Error", typeutil.HasErrorSig)
		if !ok {
			continue // No "Error" method
		}

		// Type has a `Error() string` method
		prop := None
		if embedded {
			prop |= Embedded
		}

		switch u := tn.Type().Underlying().(type) {
		case *types.Interface:
			continue // Interface type

		case *types.Struct:
			if typeutil.ZeroSized(u) {
				prop |= ZeroSized
			}

		case *types.Pointer:
			// The type is an alias of a pointer to type with an `Error() string` method.
			// This should be rare.
			prop |= PointerDef
			if typeutil.ZeroSized(u.Elem()) {
				prop |= ZeroSized
			}

		default:
			// Non-Struct error types are often value types
			prop |= NonStruct
		}

		ptrRecv := false
		if !indirect {
			// The `Error() string` method is direct or embedded without indirections
			_, ptrRecv = typeutil.HasPointerReceiver(fun.Signature())
		}

		if ptrRecv {
			// The `Error() string` has a pointer receiver
			prop |= PointerReceiver
		} else if p.HasPointerReceiverErrorMethods(tn) {
			// An error wrapping related method has a pointer receiver
			prop |= PointerMethod
		}

		// Otherwise the type has a (possibly embedded) `Error() string` method, either with value receiver
		// or the receiver type is not relevant because of indirection. We need to rely on heuristics.
		p.AddTypeProperty(tn, prop)
	}
}

// LookupMethod finds a method with the specified name in a type, checking its signature and accounting for embedding.
// Returns the method if found, whether it was found via indirection, and a boolean indicating success.
func (p pass) LookupMethod(tn *types.TypeName, name string, sigCheck func(*types.Signature) bool) (fun *types.Func, indirect, embedded, found bool) {
	obj, index, indirect := types.LookupFieldOrMethod(tn.Type(), true, p.Pkg, name)
	if obj == nil {
		return nil, false, false, false // No method with name
	}

	fun, ok := obj.(*types.Func)
	if !ok || !sigCheck(fun.Signature()) {
		return nil, false, false, false // *types.Var or wrong signature
	}

	embedded = len(index) > 1

	return fun, indirect, embedded, true
}

// errorMethods defines a list of error-related method names and functions to check their signature validity.
var errorMethods = [...]struct {
	name     string
	sigCheck func(*types.Signature) bool
}{
	{"Unwrap", typeutil.HasUnwrapSig},
	{"Is", typeutil.HasIsSig},
	{"As", typeutil.HasAsSig},
}

// HasPointerReceiverErrorMethods inspects a type for error wrapping related methods with pointer receivers.
func (p pass) HasPointerReceiverErrorMethods(tn *types.TypeName) bool {
	for _, lookup := range errorMethods {
		fun, indirect, _, ok := p.LookupMethod(tn, lookup.name, lookup.sigCheck)
		if !ok {
			continue // No such method with matching signature
		}

		if !indirect {
			// The method is direct or embedded without indirections
			if _, ptrRecv := typeutil.HasPointerReceiver(fun.Signature()); ptrRecv {
				return true
			}
		}
	}

	return false
}
