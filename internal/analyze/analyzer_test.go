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

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	testdata := analysistest.TestData()
	overrides := path.Join(testdata, "overrides.yaml")

	tests := []struct {
		name     string
		setupOpt func(*Options)
		packages []string
	}{
		{
			name: "test/a",
			setupOpt: func(o *Options) {
				do := detect.DefaultOptions()
				if err := do.ReadOverrides(overrides); err != nil {
					t.Fatalf("can't read overrides: %v", err)
				}
				o.DetectTypes = do.Analyzer()
			},
			packages: []string{"test/a"},
		},
		{
			name: "test/b",
			setupOpt: func(o *Options) {
				o.CheckIs = false
				o.DeepIsCheck = true
				o.UncheckedAssert = true
				o.CheckUnused = true
			},
			packages: []string{"test/b", "test/alias", "test/main", "test/style"},
		},
		{
			name: "test/c",
			setupOpt: func(o *Options) {
				o.StyleCheck = false
			},
			packages: []string{"test/c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			o := DefaultOptions()
			tt.setupOpt(o)
			a := o.Analyzer()

			analysistest.Run(t, testdata, a, tt.packages...)
		})
	}
}
