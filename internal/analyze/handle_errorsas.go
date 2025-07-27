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

// handleErrorsAs checks for incorrect pointer/value usage of error types passed to functions like errors.As.
func (p pass) handleErrorsAs(n *ast.CallExpr, styleCheck bool) {
	if len(n.Args) == 0 {
		return // Not interested in calls with no arguments.
	}

	// Retrieve the definition of the called function.
	fun, targetExpr, targetArgIndex := typeutil.IsErrorAs(p.TypesInfo, n)

	if fun == nil {
		return // Not an errors.As-like function.
	}

	if targetExpr != nil {
		reporter := p.GenericReporter(targetExpr, fun)

		targetType := p.TypesInfo.Types[targetExpr].Type

		// Now, check if the error type is used correctly (pointer vs. value).
		p.checkErrorUsage(targetType, reporter)

		return
	}

	if targetArgIndex < 0 {
		return // Not an errors.As-like function.
	}

	if targetArgIndex >= 0 && targetArgIndex >= len(n.Args) {
		return // Not enough arguments, e.g. called with return values of another function.
	}

	targetArg := n.Args[targetArgIndex]

	tv := p.TypesInfo.Types[targetArg]
	if !tv.IsValue() { // should not happen
		p.ReportErrorf(targetArg, "Expected value, got %#v", tv)
	}

	targetType := tv.Type

	switch t := targetType.Underlying().(type) {
	case *types.Pointer:
		// Argument is a pointer, e.g., errors.As(err, &target), which is expected.
		elemType := t.Elem()

		// The target for errors.As can be a pointer to an interface that does not
		// itself implement error (e.g., `var target interface{ Temporary() bool }`).
		// This is a valid use case for checking for specific error capabilities.
		if types.IsInterface(elemType) {
			break
		}

		// Otherwise, the pointed-to type should implement the error interface.
		// golang.org/x/tools/go/analysis/passes/errorsas checks that, too
		if !typeutil.HasErrorMethod(elemType) {
			typeName := types.TypeString(elemType, types.RelativeTo(p.Pkg))
			p.ReportRangef(targetArg, "Expected pointer to type implementing error, but %s does not. (et:arg)", typeName)

			break
		}

		reporter := p.ErrorsAsReporter(targetArg, fun)

		// Now, check if the error type is used correctly (pointer vs. value).
		p.checkErrorUsage(elemType, reporter)

		if styleCheck {
			reporter.CheckStyle(elemType)
		}

	case *types.Interface:
		// The correctness depends on the dynamic type held by the interface, which we cannot check statically.

		// Note that we don't test for `t.NumMethods() == 0`, since technically this is valid:
		//
		// 	err := struct{ error }{}
		//	var target error = &err
		//	if errors.As(err, target) { /* ... */ }

	default:
		// The argument to an errors.As-like function must be a pointer or an interface.
		var sb strings.Builder
		_ = format.Node(&sb, p.Fset, targetArg)

		p.ReportRangef(targetArg, "Target argument in %s must be a pointer or an interface, got %q (type %s). (et:arg)", fun.Name(), sb.String(), targetType)
	}
}
