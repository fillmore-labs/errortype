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
	"os/exec"
	"path"
	"sync"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	. "fillmore-labs.com/errortype/analyzer"
	"fillmore-labs.com/errortype/detect"
)

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	testdata, err := testDir()
	if err != nil {
		t.Fatal(err)
	}
	overrides := path.Join(testdata, "overrides.yaml")

	tests := [...]struct {
		name     string
		flags    map[string]string
		options  []Option
		patterns []string
	}{
		{"a", nil, []Option{
			WithDetectOptions(detect.WithOverrideFile(overrides)),
			WithNaming(true),
			WithStyleCheck(true),
		}, []string{"test/a"}},
		{"a with flags", map[string]string{
			"overrides":   overrides,
			"naming":      "true",
			"style-check": "true",
		}, nil, []string{"test/a"}},
		{"b", nil, []Option{
			WithNaming(true),
			WithCheckIs(false),
			WithDeepIsCheck(true),
			WithUncheckedAssert(true),
			WithCheckUnused(true),
			WithStyleCheck(true),
		}, []string{"test/b", "test/style", "test/main", "test/alias"}},
		{"b with flags", map[string]string{
			"naming":           "true",
			"check-is":         "false",
			"deep-is-check":    "true",
			"unchecked-assert": "true",
			"check-unused":     "true",
			"style-check":      "true",
		}, nil, []string{"test/b", "test/style", "test/main", "test/alias"}},
		{"c", nil, []Option{WithStyleCheck(false), WithNaming(true)}, []string{"test/c"}},
		{"c with flags", map[string]string{
			"style-check": "false",
			"naming":      "true",
		}, nil, []string{"test/c"}},
		{"comparable", nil, []Option{WithNotComparable(true), WithNaming(true)}, []string{"test/comparable"}},
		{"comparable with flags", map[string]string{
			"non-comparable": "true",
			"naming":         "true",
		}, nil, []string{"test/comparable"}},
		{"wrappers", nil, []Option{WithDetectOptions(detect.WithOverrideFile(overrides))}, []string{"test/wrappers"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a, err := New(tt.options...)
			if err != nil {
				t.Fatalf("Can't build analyzer: %v", err)
			}

			for name, value := range tt.flags {
				if err := a.Flags.Set(name, value); err != nil {
					t.Fatalf("can't set %s to %s: %v", name, value, err)
				}
			}

			analysistest.Run(t, testdata, a, tt.patterns...)
		})
	}
}

func TestAnalyzerFix(t *testing.T) {
	t.Parallel()

	testdata, err := testDir()
	if err != nil {
		t.Fatal(err)
	}

	tests := [...]struct {
		name     string
		flags    map[string]string
		options  []Option
		patterns []string
	}{
		{"naming", nil, []Option{WithStyleCheck(false), WithNaming(true)}, []string{"test/naming"}},
		{"naming with flags", map[string]string{
			"style-check": "false",
			"naming":      "true",
		}, nil, []string{"test/naming"}},
		{"legacy", nil, []Option{WithLegacy(true)}, []string{"test/legacy"}},
		{"legacy with flags", map[string]string{"legacy": "true"}, nil, []string{"test/legacy"}},
		{"ismethod", nil, []Option{WithLegacy(true)}, []string{"test/ismethod"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			a, err := New(tt.options...)
			if err != nil {
				t.Fatalf("Can't build analyzer: %v", err)
			}

			for name, value := range tt.flags {
				if err := a.Flags.Set(name, value); err != nil {
					t.Fatalf("can't set %s to %s: %v", name, value, err)
				}
			}

			analysistest.RunWithSuggestedFixes(t, testdata, a, tt.patterns...)
		})
	}
}

var testDir = sync.OnceValues(func() (string, error) {
	testdata := analysistest.TestData()

	cmd := exec.Command("go", "mod", "download")
	cmd.Dir = testdata

	err := cmd.Run()

	return testdata, err
})
