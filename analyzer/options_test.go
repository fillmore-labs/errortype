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
			option:   WithHeuristics(HeuristicUsage),
			expected: `"heuristics":"usage"`,
		},
		{
			name:     "WithStyleCheck",
			option:   WithStyleCheck(true),
			expected: `"style-check":true`,
		},
		{
			name:     "WithStrictIsTarget",
			option:   WithStrictIsTarget(true),
			expected: `"strict-is-target":true`,
		},
		{
			name:     "WithStrictIsMethod",
			option:   WithStrictIsMethod(true),
			expected: `"strict-is-method":true`,
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
			name:     "WithOverrides",
			option:   WithOverrides(map[string][]string{"value": {"a.ValueError", "a.OtherError"}}),
			expected: `"overrides":{"value":"a.ValueError,a.OtherError"}`,
		},
		{
			name:     "WithWrappers",
			option:   WithWrappers(map[string][]string{"is": {"a.IsError"}}),
			expected: `"wrappers":{"is":"a.IsError"}`,
		},
		{
			name: "Options",
			option: Join(
				WithTrace(regexp.MustCompile(".")),
				WithStyleCheck(true),
				WithStrictIsTarget(true),
				WithCheckUnused(true),
			),
			expected: `"options":{"trace":".","style-check":true,"strict-is-target":true,"check-unused":true}}`,
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

func TestInvalidNames(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		name   string
		option Option
		want   string
	}{
		{
			name:   "override kind",
			option: WithOverrides(map[string][]string{"pointers": {"a.PointerError"}}),
			want:   `unknown override kind "pointers"`,
		},
		{
			name:   "override type",
			option: WithOverrides(map[string][]string{"value": {"a."}}),
			want:   `invalid type name "a."`,
		},
		{
			name:   "wrapper kind",
			option: WithWrappers(map[string][]string{"iss": {"a.IsError"}}),
			want:   `unknown wrapper type: "iss"`,
		},
		{
			name:   "wrapper kind not configurable",
			option: WithWrappers(map[string][]string{"unwrap": {"a.Unwrap"}}),
			want:   `unknown wrapper type: "unwrap"`,
		},
		{
			name:   "wrapper function",
			option: WithWrappers(map[string][]string{"is": {"(a.Checker"}}),
			want:   "invalid function name",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := New(tt.option); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Expected error containing %q, got %v", tt.want, err)
			}
		})
	}
}
