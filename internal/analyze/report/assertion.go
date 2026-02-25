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

package report

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/detecttypes/result"
)

// AssertionIssue reports a configuration error or contradiction on a
// package-level compile-time error assertion.
func AssertionIssue(pass *analysis.Pass, issue result.AssertionIssue) {
	switch b := (Base{Pass: pass, Expr: issue.Target()}); issue := issue.(type) {
	case result.UnknownDirective:
		b.ReportRangef(b.Expr,
			`Unknown directive "//errortype:%s"; expected "pointer", "value", or "suppress". (et:dir)`,
			issue.Directive)

	case result.MismatchedDirective:
		if issue.Directive == result.Value {
			b.ReportRangef(b.Expr,
				`Directive "//errortype:%s" does not match pointer assertion; use "var _ error = %s" or "//errortype:%s". (et:dir)`,
				issue.Directive, b.zeroValueForm(issue.TypeName), result.Pointer)

			return
		}

		b.ReportRangef(b.Expr,
			`Directive "//errortype:%s" does not match value assertion; use "var _ error = %s" or "//errortype:%s". (et:dir)`,
			issue.Directive, b.pointerZeroForm(issue.TypeName), result.Value)

	case result.ConflictingDirective:
		msg := fmt.Sprintf(
			`Conflicting directive "//errortype:%s" for error type %q (previously "//errortype:%s"). (et:dir)`,
			issue.Directive, b.relativeNameOf(issue.TypeName), issue.RelatedDir,
		)
		b.reportWithRelated(issue.Related, msg, fmt.Sprintf(`Previous directive "//errortype:%s"`, issue.RelatedDir))

	case result.ContradictoryAssertion:
		prevKind, curKind := result.Value, result.Pointer
		if issue.FirstPtr {
			prevKind, curKind = result.Pointer, result.Value
		}

		msg := fmt.Sprintf(
			"Contradictory compile-time assertion for error type %q (%s, previously asserted as %s). (et:cfl)",
			b.relativeNameOf(issue.TypeName), curKind, prevKind,
		)
		b.reportWithRelated(issue.Related, msg, "Previously asserted as "+prevKind.String())
	}
}

func (b Base) reportWithRelated(first ast.Expr, msg, relatedMsg string) {
	var related []analysis.RelatedInformation
	if first != nil {
		related = []analysis.RelatedInformation{{
			Pos:     first.Pos(),
			End:     first.End(),
			Message: relatedMsg,
		}}
	}

	b.Report(analysis.Diagnostic{
		Pos:     b.Expr.Pos(),
		End:     b.Expr.End(),
		Message: msg,
		Related: related,
	})
}
