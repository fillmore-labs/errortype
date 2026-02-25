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
	"context"
	"go/ast"
	"runtime/trace"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/analyze/report"
	"fillmore-labs.com/errortype/internal/astutil"
	typesresult "fillmore-labs.com/errortype/internal/detecttypes/result"
	"fillmore-labs.com/errortype/internal/settings"
)

// ReportAssertionIssues reports configuration errors and contradictions on
// package-level compile-time error assertions computed by detecttypes.
func ReportAssertionIssues(ctx context.Context, pass *analysis.Pass, fileMap astutil.FileMap, issues []typesresult.AssertionIssue, opt settings.Options) {
	if len(issues) == 0 {
		return
	}

	defer trace.StartRegion(ctx, "Assertions").End()

	for _, issue := range issues {
		if file := fileMap.File(issue.Target().Pos()); skipFile(fileMap, file, pass.Analyzer.Name, opt) {
			continue
		}

		report.AssertionIssue(pass, issue)
	}
}

func skipFile(fileMap astutil.FileMap, file *ast.File, linter string, opt settings.Options) bool {
	if file == nil || !opt.Has(settings.OptionGenerated) && fileMap.IsGenerated(file) {
		return true
	}

	return astutil.HasNoLint(file.Doc, linter)
}
