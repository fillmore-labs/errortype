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

package analyze

import (
	"go/ast"

	"fillmore-labs.com/errortype/internal/knownfuncs"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// handleIsType processes type assertion functions that check if an error implements a specific type.
// It handles functions with signatures like `assert.IsType`.
func (p Pass) handleIsType(n *ast.CallExpr, methodExpr bool, ftyp knownfuncs.FuncType) {
	var typeArg int

	switch ftyp {
	case knownfuncs.IsFunc0: // suite.IsType(typeArg, targetArg, ...)
		typeArg = 0

	case knownfuncs.IsFunc1: // assert.IsType(t, typeArg, targetArg, ...)
		typeArg = 1

	default: // should not happen
		p.ReportErrorf(n, "Unsupported function type %d for IsType check", ftyp)
		return
	}

	// Receiver offset for method expressions
	if methodExpr {
		typeArg++
	}

	// Validate argument count
	if len(n.Args) < typeArg+2 {
		return // Multivalued argument or incorrect call
	}

	// Only analyze if the target implements the error interface
	targetExpr := n.Args[typeArg+1]
	if targetType, ok := p.TypesInfo.Types[targetExpr]; !ok || !typeutil.HasErrorMethod(targetType.Type) {
		return
	}

	typeExpr := n.Args[typeArg]

	assertedType, ok := p.TypesInfo.Types[typeExpr]
	if !ok { // should not happen
		return
	}

	// Check if the type is correctly asserted
	p.checkAssert(assertedType.Type, typeExpr)
}
