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

	"fillmore-labs.com/errortype/internal/typeutil"
)

// handleTypeSwitch checks for incorrect pointer/value usage of error types in type switch cases.
func (p pass) handleTypeSwitch(n *ast.TypeSwitchStmt) {
	// expr must be of interface type, but we don't check
	expr, ok := getTypeSwitchExpr(n)
	if !ok { // should not happen
		p.ReportErrorf(n, "Cannot analyze type switch: unable to determine switch expression")

		return
	}

	// We are only interested in switches on an error type.
	if tv, ok := p.TypesInfo.Types[expr]; !ok || !typeutil.HasErrorMethod(tv.Type) {
		return // Not a switch on an error type.
	}

	// Iterate through all "case" clauses in the switch statement.
	for _, stmt := range n.Body.List {
		clause, ok := stmt.(*ast.CaseClause)
		if !ok { // should not happen
			p.ReportErrorf(stmt, "Expected a case clause in type switch, but got %T", stmt)

			continue
		}

		// Check each type in the case clause (e.g., "case T1, T2:").
		for _, caseExpr := range clause.List {
			if caseExpr == nil {
				continue // ignore the "default:" case
			}

			caseType := p.TypesInfo.Types[caseExpr]
			if caseType.IsNil() {
				continue // ignore the "nil:" case
			}

			if !caseType.IsType() { // should not happen
				p.ReportErrorf(caseExpr, "Expected a type in case clause, but got %v", caseType)

				continue
			}

			// Perform the pointer-vs-value analysis on the case type.
			p.checkErrorUsage(caseType.Type, p.SwitchReporter(caseExpr))
		}
	}
}

// getTypeSwitchExpr extracts the expression being type-switched on from an *ast.TypeSwitchStmt.
// It handles both "switch x := y.(type)" and "switch y.(type)" forms.
func getTypeSwitchExpr(n *ast.TypeSwitchStmt) (ast.Expr, bool) {
	var typeAssert *ast.TypeAssertExpr

	switch s := n.Assign.(type) {
	case *ast.AssignStmt: // switch x := y.(type)
		if len(s.Rhs) > 0 {
			typeAssert, _ = s.Rhs[0].(*ast.TypeAssertExpr)
		}
	case *ast.ExprStmt: // switch y.(type)
		typeAssert, _ = s.X.(*ast.TypeAssertExpr)
	}

	// A valid type switch must have a type assertion of the form ".(type)".
	// If not, the source is not valid Go, but we check defensively.
	if typeAssert == nil || typeAssert.Type != nil {
		return nil, false
	}

	return typeAssert.X, true
}
