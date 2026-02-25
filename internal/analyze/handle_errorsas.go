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
	"go/format"
	"go/types"
	"strings"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// handleErrorsAs checks for incorrect pointer/value usage of error types passed to functions like `errors.As`.
func (p Pass) handleErrorsAs(n *ast.CallExpr, ex ast.Expr, targetArgIndex int) {
	if targetArgIndex >= len(n.Args) {
		return // Not enough arguments, e.g. called with return values of another function.
	}

	targetArg := n.Args[targetArgIndex]

	tv, ok := p.TypesInfo.Types[targetArg]
	if !ok || !tv.IsValue() { // should not happen
		p.ReportErrorf(targetArg, "Expected argument value, got %#v", tv)
	}

	switch targetType := tv.Type.Underlying().(type) {
	case *types.Pointer:
		// Argument is a pointer, e.g., errors.As(err, &target), which is expected.
		elemType := targetType.Elem()

		// The target for errors.As can be a pointer to an interface that is not
		// required to implement `error` (e.g., `var target interface{ Temporary() bool }`).
		// This is a valid use case for checking for specific error capabilities.
		if types.IsInterface(elemType) {
			return
		}

		// The pointed-to type must implement the error interface.
		if !p.implementsError(elemType, targetArg) {
			return
		}

		// Now, check if the error type is used correctly (pointer vs. value).
		p.checkErrorsAs(elemType, targetArg, ex)

	case *types.Interface:
		// The correctness depends on the dynamic type held by the interface, which we cannot check statically.
		if targetType.NumMethods() == 0 {
			return
		}

		// The dynamic type of the interface must be a non-nil pointer
		// that points either to an interface or a type implementing `error`
		//
		// While the interface itself does not have to be `any`, everything
		// else is rare and error-prone:
		//
		//	type TestError struct{ ... }
		//	func (TestError) Error() string { ... }
		//
		//	var err error = &TestError{ ... }
		//	if errors.As(TestError{...}, err) { ... }
		typeName := types.TypeString(tv.Type, types.RelativeTo(p.Pkg))
		p.ReportRangef(targetArg, `Expected pointer or "any" interface, but %q is a non-empty interface. (et:arg+)`, typeName)

	default:
		// The argument to an `errors.As`-like function must be a pointer or an interface.
		funName := "<invalid>"
		if sb := (strings.Builder{}); format.Node(&sb, p.Fset, n.Fun) == nil {
			funName = sb.String()
		}

		target := "<invalid>"
		if sb := (strings.Builder{}); format.Node(&sb, p.Fset, targetArg) == nil {
			target = sb.String()
		}

		typeName := types.TypeString(tv.Type, types.RelativeTo(p.Pkg))
		p.ReportRangef(targetArg, `Target argument of %s must be a pointer or interface, got %q (type %s). (et:arg)`, funName, target, typeName)
	}
}

func (p Pass) implementsError(elemType types.Type, targetArg ast.Expr) bool {
	if typeutil.HasErrorMethod(elemType) {
		return true
	}

	typeName := types.TypeString(elemType, types.RelativeTo(p.Pkg))
	p.ReportRangef(targetArg, `Expected pointer to a type implementing "error", but %q does not. (et:arg)`, typeName)

	return false
}
