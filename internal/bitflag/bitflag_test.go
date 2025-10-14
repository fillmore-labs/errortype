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

package bitflag_test

import (
	"testing"

	. "fillmore-labs.com/errortype/internal/bitflag"
)

func TestToString(t *testing.T) {
	t.Parallel()

	names := []string{
		0: "One",
		1: "Two",
		2: "Four",
	}

	tests := []struct {
		name  string
		value uint8
		want  string
	}{
		{name: "Zero value", value: 0, want: "Zero"},
		{name: "Single value", value: 1, want: "One"},
		{name: "Two values", value: 1 | 4, want: "Four, One"},
		{name: "Three values", value: 1 | 2 | 4, want: "Four, Two, One"},
		{name: "Unknown value", value: 8, want: "Unknown(3)"},
		{name: "Mixed values", value: 1 | 8, want: "Unknown(3), One"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ToString(tt.value, names, "Zero"); got != tt.want {
				t.Errorf("ToString() = %q, want %q", got, tt.want)
			}
		})
	}
}
