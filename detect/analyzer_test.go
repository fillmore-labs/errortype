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

package detect_test

import (
	"fmt"
	"go/ast"
	"go/types"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	. "fillmore-labs.com/errortype/detect"
	"fillmore-labs.com/errortype/detect/result"
	"fillmore-labs.com/errortype/internal/typeutil"
)

func TestDetectAnalyzer(t *testing.T) {
	t.Parallel()

	dir := analysistest.TestData()
	overrides := filepath.Join(dir, "overrides.yaml")

	tests := [...]struct {
		name        string
		newAnalyzer func() *analysis.Analyzer
		pkg         string
	}{
		{"detect", func() *analysis.Analyzer {
			t.Helper()

			return New(WithOverrideFile(overrides))
		}, "./a/c"},
		{"wrappers", func() *analysis.Analyzer {
			t.Helper()

			return New(WithOverrideFile(overrides))
		}, "./wrappers/..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			analysistest.Run(t, dir, tt.newAnalyzer(), tt.pkg)
		})
	}
}

func TestDetectAnalyzerResults(t *testing.T) {
	t.Parallel()

	dir := analysistest.TestData()
	overrides := filepath.Join(dir, "overrides.yaml")

	tests := [...]struct {
		name        string
		newAnalyzer func() *analysis.Analyzer
		pkg         string
	}{
		{"errortypes", func() *analysis.Analyzer {
			t.Helper()

			d := New(WithOverrideFile(overrides))

			return newTestAnalyzer(d)
		}, "test/a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			analysistest.Run(t, dir, tt.newAnalyzer(), tt.pkg)
		})
	}
}

type testFact struct{ result.ErrorType }

func newTestAnalyzer(d *analysis.Analyzer) *analysis.Analyzer {
	testAnalyzer := &analysis.Analyzer{
		Name: "testanalyzer",
		Doc:  "consumes results from detect.Analyzer for testing",
		Run: func(ap *analysis.Pass) (any, error) {
			return run(ap, d)
		},
		Requires:  []*analysis.Analyzer{inspect.Analyzer, d},
		FactTypes: []analysis.Fact{(*testFact)(nil)},
	}

	return testAnalyzer
}

func run(ap *analysis.Pass, d *analysis.Analyzer) (any, error) {
	in, ok := ap.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, fmt.Errorf("testanalyzer: analyzer %s result missing", inspect.Analyzer.Name)
	}

	res, ok := ap.ResultOf[d].(result.Result)
	if !ok {
		return nil, fmt.Errorf("testanalyzer: analyzer %s result missing", d.Name)
	}

	errorMap := make(result.ErrorTypes, len(res.Types))
	for name, errorType := range res.Types {
		errorMap[name] = errorType

		if ap.Pkg == name.Pkg() {
			ap.ExportObjectFact(name, &testFact{errorType})
		}
	}

	errorInterface := types.Universe.Lookup("error").Type().Underlying().(*types.Interface)

	for caseClause := range inspector.All[*ast.CaseClause](in) {
		if len(caseClause.Body) != 1 {
			continue
		}

		assignStmt, ok := caseClause.Body[0].(*ast.AssignStmt)
		if !ok {
			continue
		}

		for _, rhs := range assignStmt.Rhs {
			tv, ok := ap.TypesInfo.Types[rhs]
			if !ok || tv.IsNil() {
				continue
			}
			t := tv.Type

			if !types.Implements(t, errorInterface) {
				continue
			}

			msg := message(t, errorMap)
			ap.ReportRangef(rhs, "Type %q %s", t.String(), msg)
		}
	}

	return nil, nil
}

func message(t types.Type, errorMap map[*types.TypeName]result.ErrorType) string {
	tn, _, ok := typeutil.TypeNameOf(t)
	if !ok {
		return "NOT A NAMED TYPE"
	}

	typ, ok := errorMap[tn]
	if !ok {
		return "NOT IN RESULTS"
	}

	switch typ & result.ExpectedMask {
	case result.Undecided:
		return "UNDECIDED"

	case result.Pointer:
		return "POINTER"

	case result.Value:
		return "VALUE"

	case result.Suppress:
		return "SUPPRESS"

	default:
		return "ERROR"
	}
}
