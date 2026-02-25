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

import (
	"go/ast"
	"go/token"
	"go/types"
)

// BuiltinNew checks whether the given call expression is a call to the built-in [new] function.
func BuiltinNew(info *types.Info, call *ast.CallExpr) bool {
	if len(call.Args) != 1 || call.Ellipsis != token.NoPos {
		return false // some function
	}

	fun, ok := ast.Unparen(call.Fun).(*ast.Ident)
	if !ok || fun.Name != "new" {
		return false // not new(...)
	}

	f, ok := info.Uses[fun]

	return ok && f == builtinNew
}

var builtinNew = types.Universe.Lookup("new").(*types.Builtin)
