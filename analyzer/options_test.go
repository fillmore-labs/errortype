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
	"regexp"
	"strings"
	"testing"

	. "fillmore-labs.com/errortype/analyzer"
	"fillmore-labs.com/errortype/detect"
)

func TestLogValue(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		name     string
		option   Option
		expected string
	}{
		{
			name:     "WithDetectTypes",
			option:   WithDetectOptions(detect.WithHeuristics(detect.HeuristicUsage)),
			expected: `"detect-options":{"heuristics":"usage"}`,
		},
		{
			name:     "WithStyleCheck",
			option:   WithStyleCheck(true),
			expected: `"style-check":true`,
		},
		{
			name:     "WithCheckIs",
			option:   WithCheckIs(false),
			expected: `"check-is":false`,
		},
		{
			name:     "WithDeepIsCheck",
			option:   WithDeepIsCheck(true),
			expected: `"deep-is-check":true`,
		},
		{
			name:     "WithUncheckedAssert",
			option:   WithUncheckedAssert(false),
			expected: `"unchecked-assert":false`,
		},
		{
			name:     "WithCheckUnused",
			option:   WithCheckUnused(true),
			expected: `"check-unused":true`,
		},
		{
			name: "Options",
			option: Join(
				WithDetectOptions(detect.WithTrace(regexp.MustCompile("."))),
				WithStyleCheck(true),
				WithCheckIs(false),
				WithCheckUnused(true),
			),
			expected: `"options":{"detect-options":{"trace":"."},"style-check":true,"check-is":false,"check-unused":true}}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var sb strings.Builder
			logger := slog.New(slog.NewJSONHandler(&sb, nil))
			logger.LogAttrs(t.Context(), slog.LevelInfo, "test", tt.option.LogAttr())

			if got := sb.String(); !strings.Contains(got, tt.expected) {
				t.Errorf("Expected %s to contain %s", got, tt.expected)
			}
		})
	}
}
