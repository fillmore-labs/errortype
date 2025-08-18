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
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	. "fillmore-labs.com/errortype/internal/analyze"
	"fillmore-labs.com/errortype/internal/detect"
)

func TestAnalyzerA(t *testing.T) {
	t.Parallel()

	testdata := analysistest.TestData()

	d := detect.New()
	a := New(WithDetectTypes(d))

	if err := d.Flags.Set("overrides", path.Join(testdata, "overrides.yaml")); err != nil {
		t.Fatal("can't set override file", err)
	}

	analysistest.Run(t, testdata, a, "test/a")
}

func TestAnalyzerB(t *testing.T) {
	t.Parallel()

	testdata := analysistest.TestData()

	d := detect.New()
	a := New(WithDetectTypes(d))

	if err := a.Flags.Set("check-is", "false"); err != nil {
		t.Fatal("can't set check-is", err)
	}

	if err := a.Flags.Set("deep-is-check", "true"); err != nil {
		t.Fatal("can't set deep-is-check", err)
	}

	analysistest.Run(t, testdata, a, "test/b", "test/alias")
}

func TestAnalyzerC(t *testing.T) {
	t.Parallel()

	testdata := analysistest.TestData()

	d := detect.New()
	a := New(WithDetectTypes(d))

	if err := a.Flags.Set("stylecheck", "false"); err != nil {
		t.Fatal("can't set stylecheck", err)
	}

	analysistest.Run(t, testdata, a, "test/c")
}
