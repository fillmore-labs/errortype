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
	overrides := path.Join(testdata, "overrides.yaml")

	d, err := detect.New(detect.WithOverrideFile(overrides))
	if err != nil {
		t.Fatalf("Can't build detect analyzer: %v", err)
	}

	tests := [...]struct {
		name     string
		patterns []string
		options  []Option
		flags    func(*flag.FlagSet)
	}{
		{"a", []string{"test/a"}, []Option{WithDetectTypes(d), WithNaming(true), WithStyleCheck(true)}, nil},
		{
			"a with flags", []string{"test/a"}, nil, setFlags(t, map[string]string{
				"overrides":   overrides,
				"naming":      "true",
				"style-check": "true",
			}),
		},
		{"b", []string{"test/b", "test/style", "test/main", "test/alias"}, []Option{WithCheckIs(false), WithDeepIsCheck(true), WithUncheckedAssert(true), WithCheckUnused(true), WithNaming(true), WithStyleCheck(true)}, nil},
		{
			"b with flags", []string{"test/b", "test/style", "test/main", "test/alias"}, nil, setFlags(t, map[string]string{
				"check-is":         "false",
				"deep-is-check":    "true",
				"unchecked-assert": "true",
				"check-unused":     "true",
				"naming":           "true",
				"style-check":      "true",
			}),
		},
		{"c", []string{"test/c"}, []Option{WithStyleCheck(false), WithNaming(true)}, nil},
		{
			"c with flags", []string{"test/c"}, nil, setFlags(t, map[string]string{
				"style-check": "false",
				"naming":      "true",
			}),
		},
		{"wrappers", []string{"test/wrappers"}, []Option{WithDetectTypes(d)}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a, err := New(tt.options...)
			if err != nil {
				t.Fatalf("Can't build analyzer: %v", err)
			}

			if tt.flags != nil {
				tt.flags(&a.Flags)
			}

			analysistest.Run(t, testdata, a, tt.patterns...)
		})
	}
}

func setFlags(tb testing.TB, flags map[string]string) func(*flag.FlagSet) { //nolint:thelper
	return func(fs *flag.FlagSet) {
		tb.Helper()

		for name, value := range flags {
			if err := fs.Set(name, value); err != nil {
				tb.Fatalf("can't set %s to %s: %v", name, value, err)
			}
		}
	}
}
