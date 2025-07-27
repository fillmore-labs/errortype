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
	"fillmore-labs.com/errortype/internal/typeutil"
)

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	if err := typeutil.HasGo(); err != nil {
		t.Skipf("Go not available: %s", err)
	}

	testdata := analysistest.TestData()

	d := detect.New()
	a := New(WithDetectTypes(d))

	if err := d.Flags.Set("overrides", path.Join(testdata, "overrides.yaml")); err != nil {
		t.Fatal("can't set override file", err)
	}

	analysistest.Run(t, testdata, a, "test/...")
}
