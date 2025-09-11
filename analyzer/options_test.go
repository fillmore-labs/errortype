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
	"log/slog"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"

	. "fillmore-labs.com/errortype/analyzer"
)

func TestLogValue(t *testing.T) {
	t.Parallel()

	testAnalyzer := &analysis.Analyzer{Name: "test-detect"}

	testCases := []struct {
		name     string
		option   Option
		expected string
	}{
		{
			name:     "WithDetectTypes",
			option:   WithDetectTypes(testAnalyzer),
			expected: `"detect":"test-detect"`,
		},
		{
			name:     "WithStyleCheck",
			option:   WithStyleCheck(true),
			expected: `"styleCheck":true`,
		},
		{
			name:     "WithCheckIs",
			option:   WithCheckIs(false),
			expected: `"checkIs":false`,
		},
		{
			name:     "WithDeepIsCheck",
			option:   WithDeepIsCheck(true),
			expected: `"deepIsCheck":true`,
		},
		{
			name:     "WithUncheckedAssert",
			option:   WithUncheckedAssert(false),
			expected: `"uncheckedAssert":false`,
		},
		{
			name: "Options",
			option: Options{
				WithDetectTypes(testAnalyzer),
				WithStyleCheck(true),
				WithCheckIs(false),
			},
			expected: `"options":{"detect":"test-detect","styleCheck":true,"checkIs":false}}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var sb strings.Builder
			logger := slog.New(slog.NewJSONHandler(&sb, nil))
			logger.LogAttrs(t.Context(), slog.LevelInfo, "test", tt.option.SlogAttr())

			got := sb.String()
			if !strings.Contains(got, tt.expected) {
				t.Errorf("Expected %s to contain %s", got, tt.expected)
			}
		})
	}
}
