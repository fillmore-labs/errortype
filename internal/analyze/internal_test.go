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
	"fmt"
	"go/ast"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"fillmore-labs.com/errortype/detect/result"
)

func TestAnalyzerInternal(t *testing.T) {
	t.Parallel()

	testdata := analysistest.TestData()

	testAnalyzer := &analysis.Analyzer{
		Name:     "testanalyzer",
		Doc:      "tests internal errors",
		Run:      run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}

	analysistest.Run(t, testdata, testAnalyzer, "test/internal")
}

func run(pass *analysis.Pass) (any, error) {
	in, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, fmt.Errorf("testanalyzer: %s reuslt missing", inspect.Analyzer.Name)
	}

	p := NewPass(pass, result.Result{}, DefaultOptions)

	var count int
	for n := range inspector.All[*ast.ValueSpec](in) {
		count++
		p.ReportErrorf(n, "Error %d", count)
	}

	return any(nil), nil
}
