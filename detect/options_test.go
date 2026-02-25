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

package detect_test

import (
	"log/slog"
	"regexp"
	"strings"
	"testing"

	. "fillmore-labs.com/errortype/detect"
	"fillmore-labs.com/errortype/detect/result"
)

func TestLogValue(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		name     string
		option   Option
		expected string
	}{
		{
			name:     "WithOverrides",
			option:   WithOverrides(map[result.ErrorType][]string{result.Pointer: {"pkg.MyType"}}),
			expected: `"overrides":{"pointer":"pkg.MyType"}`,
		},
		{
			name: "WithOverrides Multiple",
			option: WithOverrides(map[result.ErrorType][]string{
				result.Suppress: {"pkg.Suppress"},
				result.Pointer:  {"pkg.Pointer"},
				result.Value:    {"pkg.Value"},
			}),
			expected: `"overrides":{"pointer":"pkg.Pointer","value":"pkg.Value","suppressed":"pkg.Suppress"}`,
		},
		{
			name:     "WithWrappers",
			option:   WithWrappers(map[result.WrapperType][]string{result.WrapperAsType: {"pkg.MyAsType"}}),
			expected: `"wrappers":{"astype":"pkg.MyAsType"}`,
		},
		{
			name: "WithWrappers Multiple",
			option: WithWrappers(map[result.WrapperType][]string{
				result.FuncAssert:    {"pkg.Assert"},
				result.WrapperAsType: {"pkg.AsType"},
				result.WrapperAs:     {"pkg.As"},
				result.WrapperIs:     {"pkg.Is"},
				result.FuncEqual:     {"pkg.Equal"},
				result.FuncIsType:    {"pkg.IsType"},
			}),
			expected: `"wrappers":{"is":"pkg.Is","as":"pkg.As","astype":"pkg.AsType","istype":"pkg.IsType","equal":"pkg.Equal","assert":"pkg.Assert"}`,
		},
		{
			name:     "WithHeuristics - None",
			option:   WithHeuristics(),
			expected: `"heuristics":""`,
		},
		{
			name:     "WithHeuristics - Off",
			option:   WithHeuristics(HeuristicOff),
			expected: `"heuristics":"off"`,
		},
		{
			name:     "WithHeuristics - Usage",
			option:   WithHeuristics(HeuristicUsage),
			expected: `"heuristics":"usage"`,
		},
		{
			name:     "WithHeuristics - All",
			option:   WithHeuristics(HeuristicVar, HeuristicUsage, HeuristicReceivers),
			expected: `"heuristics":"var,usage,receivers"`,
		},
		{
			name:     "WithTrace",
			option:   WithTrace(regexp.MustCompile(".*")),
			expected: `"trace":".*"`,
		},
		{
			name: "Options",
			option: Join(
				WithOverrides(map[result.ErrorType][]string{result.Value: {"pkg.MyType"}}),
				WithHeuristics(HeuristicUsage),
				WithTrace(regexp.MustCompile(".*")),
			),
			expected: `"options":{"overrides":{"value":"pkg.MyType"},"heuristics":"usage","trace":".*"}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var sb strings.Builder
			logger := slog.New(slog.NewJSONHandler(&sb, nil))
			logger.LogAttrs(t.Context(), slog.LevelInfo, "test", tt.option.LogAttr())

			if got := sb.String(); !strings.Contains(got, tt.expected) {
				t.Errorf("Expected log output %s to contain %s", got, tt.expected)
			}
		})
	}
}
