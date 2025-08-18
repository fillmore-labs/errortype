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
	"errors"
	"go/ast"
	"go/types"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	. "fillmore-labs.com/errortype/internal/detect"
	"fillmore-labs.com/errortype/internal/errortypes"
)

func TestExclusionsAnalyzer(t *testing.T) {
	t.Parallel()

	d := New()

	dir := analysistest.TestData()

	if err := d.Flags.Set("overrides", filepath.Join(dir, "overrides.yaml")); err != nil {
		t.Fatalf("can't set overrides flag: %v", err)
	}

	testAnalyzer := &analysis.Analyzer{
		Name: "testanalyzer",
		Doc:  "consumes results from detect.Analyzer for testing",
		Run: func(ap *analysis.Pass) (any, error) {
			return run(ap, d)
		},
		Requires: []*analysis.Analyzer{inspect.Analyzer, d},
	}

	tests := []struct {
		name     string
		analyzer *analysis.Analyzer
		pkg      string
	}{
		{"errortypes", testAnalyzer, "test/a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			analysistest.Run(t, dir, tt.analyzer, tt.pkg)
		})
	}
}

var (
	// ErrNoInspectorResult is returned when the ast inspector is missing.
	ErrNoInspectorResult = errors.New("testanalyzer: inspector result missing")

	// ErrNoDetecttypesResult is returned when the detecttypes result is missing.
	ErrNoDetecttypesResult = errors.New("testanalyzer: detecttypes result missing")
)

func run(ap *analysis.Pass, d *analysis.Analyzer) (any, error) {
	in, ok := ap.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, ErrNoInspectorResult
	}

	res, ok := ap.ResultOf[d].(Result)
	if !ok {
		return nil, ErrNoDetecttypesResult
	}

	errorMap := make(map[*types.TypeName]errortypes.ErrorType)
	for _, info := range res.Types {
		errorMap[info.TypeName] = info.ErrorType
	}

	for returnStmt := range inspector.All[*ast.ReturnStmt](in) {
		for _, result := range returnStmt.Results {
			tv, ok := ap.TypesInfo.Types[result]
			if !ok || tv.IsNil() {
				continue
			}

			t := tv.Type
			if p, ok := t.(*types.Pointer); ok {
				t = p.Elem()
			}

			msg := message(t, errorMap)

			ap.ReportRangef(result, "Type %q %s", t.String(), msg)
		}
	}

	return any(nil), nil
}

func message(t types.Type, errorMap map[*types.TypeName]errortypes.ErrorType) string {
	var tn *types.TypeName
	switch t := t.(type) {
	case *types.Named:
		tn = t.Obj()

	case *types.Alias:
		tn = t.Obj()

	default:
		return "NOT A NAMED TYPE"
	}

	typ, ok := errorMap[tn]
	if !ok {
		return "NOT IN RESULTS"
	}

	switch typ {
	case errortypes.PointerType:
		return "POINTER"

	case errortypes.ValueType:
		return "VALUE"

	case errortypes.SuppressType:
		return "SUPPRESS"

	default:
		return "ERROR"
	}
}
