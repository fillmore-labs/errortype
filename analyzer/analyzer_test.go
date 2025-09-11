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

package analyzer_test

import (
	"flag"
	"path"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	. "fillmore-labs.com/errortype/analyzer"
	"fillmore-labs.com/errortype/detect"
)

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	testdata := analysistest.TestData()

	detectTypes := func() Option {
		t.Helper()

		overridefile := path.Join(testdata, "overrides.yaml")

		d := detect.New()

		if err := d.Flags.Set("overrides", overridefile); err != nil {
			t.Fatal("can't set override file", err)
		}

		return WithDetectTypes(d)
	}

	tests := []struct {
		name     string
		patterns []string
		options  Options
		flags    func(*flag.FlagSet)
	}{
		{"a with flags", []string{"test/a"}, []Option{detectTypes()}, nil},
		{"b", []string{"test/b", "test/style"}, []Option{WithCheckIs(false), WithDeepIsCheck(true), WithUncheckedAssert(true)}, nil},
		{"b with flags", []string{"test/b", "test/style"}, nil, func(f *flag.FlagSet) {
			t.Helper()

			if err := f.Set("check-is", "false"); err != nil {
				t.Fatal("can't set check-is", err)
			}

			if err := f.Set("deep-is-check", "true"); err != nil {
				t.Fatal("can't set deep-is-check", err)
			}

			if err := f.Set("unchecked-assert", "true"); err != nil {
				t.Fatal("can't set unchecked-assert", err)
			}
		}},
		{"c", []string{"test/c"}, []Option{WithStyleCheck(false)}, nil},
		{"c with flags", []string{"test/c"}, nil, func(f *flag.FlagSet) {
			t.Helper()

			if err := f.Set("style-check", "false"); err != nil {
				t.Fatal("can't set style-check", err)
			}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a := New(tt.options...)

			if tt.flags != nil {
				tt.flags(&a.Flags)
			}

			analysistest.Run(t, testdata, a, tt.patterns...)
		})
	}
}
