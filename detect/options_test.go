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
	"strings"
	"testing"

	. "fillmore-labs.com/errortype/detect"
)

func TestLogValue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		option   Option
		expected string
	}{
		{
			name:     "WithOverrides",
			option:   WithOverrides(map[Override][]string{OverridePointer: {"pkg.MyType"}}),
			expected: `"overrides":{"pointer":"pkg.MyType"}`,
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
			option:   WithHeuristics(HeuristicUsage, HeuristicReceivers),
			expected: `"heuristics":"usage,receivers"`,
		},
		{
			name:     "WithTrace",
			option:   WithTrace(true),
			expected: `"trace":true`,
		},
		{
			name: "Options",
			option: Options{
				WithOverrides(map[Override][]string{OverrideValue: {"pkg.MyType"}}),
				WithHeuristics(HeuristicUsage),
				WithTrace(true),
			},
			expected: `"options":{"overrides":{"value":"pkg.MyType"},"heuristics":"usage","trace":true}`,
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
				t.Errorf("Expected log output %s to contain %s", got, tt.expected)
			}
		})
	}
}
