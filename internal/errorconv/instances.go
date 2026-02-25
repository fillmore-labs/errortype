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

package errorconv

import (
	"go/ast"
	"go/types"
	"iter"

	"fillmore-labs.com/errortype/internal/errorlike"
)

// TypeArg represents a concrete type argument instantiated for a type parameter
// whose constraint has an `Error() string` method.
type TypeArg struct {
	Ident *ast.Ident
	Obj   types.Object
	Type  types.Type
	Index int
}

// TypeArgs yields each non-interface type argument in info.Instances whose
// corresponding type parameter constraint has an `Error() string` method.
// It covers generic functions, generic methods, generic defined types, and
// generic type aliases, for both explicit and inferred instantiations.
func TypeArgs(info *types.Info) iter.Seq[TypeArg] {
	return func(yield func(TypeArg) bool) {
		for id, inst := range info.Instances {
			obj, ok := info.Uses[id]
			if !ok {
				continue
			}

			t, ok := obj.Type().(interface{ TypeParams() *types.TypeParamList })
			if !ok {
				continue
			}

			tparams, targs := t.TypeParams(), inst.TypeArgs
			for i := range min(tparams.Len(), targs.Len()) {
				if !errorlike.IsInterface(tparams.At(i).Constraint()) {
					continue
				}

				argType := targs.At(i)
				if types.IsInterface(argType) {
					continue
				}

				if !yield(TypeArg{Ident: id, Obj: obj, Type: argType, Index: i}) {
					return
				}
			}
		}
	}
}
