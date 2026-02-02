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
	"fillmore-labs.com/errortype/facts"
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

func newTestAnalyzer(d *analysis.Analyzer) *analysis.Analyzer {
	testAnalyzer := &analysis.Analyzer{
		Name: "testanalyzer",
		Doc:  "consumes results from detect.Analyzer for testing",
		Run: func(ap *analysis.Pass) (any, error) {
			return run(ap, d)
		},
		Requires: []*analysis.Analyzer{inspect.Analyzer, d},
	}

	return testAnalyzer
}

func run(ap *analysis.Pass, d *analysis.Analyzer) (any, error) {
	in, ok := ap.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, fmt.Errorf("testanalyzer: analyzer %s result missing", inspect.Analyzer.Name)
	}

	res, ok := ap.ResultOf[d].(facts.Result)
	if !ok {
		return nil, fmt.Errorf("testanalyzer: analyzer %s result missing", d.Name)
	}

	errorMap := make(map[*types.TypeName]facts.ErrorFact, len(res.Types))
	for _, info := range res.Types {
		errorMap[info.TypeName] = info.ErrorType
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

		for _, result := range assignStmt.Rhs {
			tv, ok := ap.TypesInfo.Types[result]
			if !ok || tv.IsNil() {
				continue
			}
			t := tv.Type

			if !types.Implements(t, errorInterface) {
				continue
			}

			msg := message(t, errorMap)
			ap.ReportRangef(result, "Type %q %s", t.String(), msg)
		}
	}

	return any(nil), nil
}

func message(t types.Type, errorMap map[*types.TypeName]facts.ErrorFact) string {
	tn, _, ok := typeutil.TypeNameOf(t)
	if !ok {
		return "NOT A NAMED TYPE"
	}

	typ, ok := errorMap[tn]
	if !ok {
		return "NOT IN RESULTS"
	}

	switch typ & facts.ExpectedMask {
	case facts.UndecidedType:
		return "UNDECIDED"

	case facts.PointerType:
		return "POINTER"

	case facts.ValueType:
		return "VALUE"

	case facts.SuppressType:
		return "SUPPRESS"

	default:
		return "ERROR"
	}
}
