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

package analyze_test

import (
	"path"
	"reflect"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"fillmore-labs.com/errortype/detect"
	. "fillmore-labs.com/errortype/internal/analyze"
)

func TestAnalyzerA(t *testing.T) {
	t.Parallel()

	o := DefaultOptions()
	a := newAnalyzer(o)

	testdata := analysistest.TestData()

	d := o.DetectTypes
	if err := d.Flags.Set("overrides", path.Join(testdata, "overrides.yaml")); err != nil {
		t.Fatal("can't set override file", err)
	}

	analysistest.Run(t, testdata, a, "test/a")
}

func TestAnalyzerB(t *testing.T) {
	t.Parallel()

	o := DefaultOptions()
	o.CheckIs = false
	o.DeepIsCheck = true
	o.UncheckedAssert = true
	a := newAnalyzer(o)

	testdata := analysistest.TestData()

	analysistest.Run(t, testdata, a, "test/b", "test/alias", "test/main", "test/style")
}

func TestAnalyzerC(t *testing.T) {
	t.Parallel()

	testdata := analysistest.TestData()

	o := DefaultOptions()
	o.StyleCheck = false
	a := newAnalyzer(o)

	analysistest.Run(t, testdata, a, "test/c")
}

func newAnalyzer(o *Options) *analysis.Analyzer {
	if o.DetectTypes == nil {
		o.DetectTypes = detect.New()
	}

	return &analysis.Analyzer{
		Name:       "errortype",
		Doc:        "errortype is a Go static analysis tool that helps prevent subtle bugs in error handling.",
		Run:        o.Run,
		Requires:   []*analysis.Analyzer{inspect.Analyzer, o.DetectTypes},
		ResultType: reflect.TypeFor[Result](),
	}
}
