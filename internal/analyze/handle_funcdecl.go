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

package analyze

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/astutil"
)

// handleFuncDecl detects legacy error query helper functions that perform direct type assertions
// on error interfaces and suggests to use errors.As or errors.AsType instead, e.g.:
//
//	func IsParseError(err error) bool {
//		_, ok := err.(*net.ParseError)
//		return ok
//	}
func (p Pass) handleFuncDecl(f *ast.FuncDecl, file *ast.File) {
	if !p.Has(OptionLegacy) {
		return
	}

	q, ok := p.matchAssertionQuery(f.Type, f.Body)
	if !ok {
		return
	}

	if _, suppressed := astutil.AnalyzeComments(f.Doc, "", p.Analyzer.Name); suppressed {
		return
	}

	var fixes []analysis.SuggestedFix
	if edits := q.legacyFix(p.TypesInfo, p.Fset, file); len(edits) > 0 {
		fixes = []analysis.SuggestedFix{{
			Message:   "Replace with errors.As/AsType",
			TextEdits: edits,
		}}
	}

	p.Report(analysis.Diagnostic{
		Pos:            f.Type.Pos(),
		End:            f.Type.End(),
		Message:        fmt.Sprintf("Legacy error assertion query %q does not account for wrapped errors; use errors.As/AsType (et:lgc)", f.Name.Name),
		SuggestedFixes: fixes,
	})
}
