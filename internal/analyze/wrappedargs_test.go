// Copyright 2026 Oliver Eikemeier. All Rights Reserved.
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
	"errors"
	"fmt"
	"slices"
	"strconv"
	"testing"

	. "fillmore-labs.com/errortype/internal/analyze"
)

func TestWrappedArgs(t *testing.T) {
	t.Parallel()

	tests := [...]struct {
		name    string
		format  string
		want    []int
		numArgs int
	}{
		{name: "empty", format: "", numArgs: 0, want: nil},
		{name: "no verbs", format: "no verbs", numArgs: 1, want: nil},
		{name: "simple", format: "%w", numArgs: 1, want: []int{0}},
		{name: "not wrapping", format: "%v", numArgs: 1, want: nil},
		{name: "second operand", format: "%s: %w", numArgs: 2, want: []int{1}},
		{name: "percent", format: "100%% %w", numArgs: 1, want: []int{0}},
		{name: "percent w", format: "%%w", numArgs: 1, want: nil},
		{name: "triple percent", format: "%%%w", numArgs: 1, want: []int{0}},
		{name: "flagged percent", format: "%+% %w", numArgs: 1, want: []int{0}},
		{name: "flags", format: "%+w", numArgs: 1, want: []int{0}},
		{name: "all flags", format: "%#0+- w", numArgs: 1, want: []int{0}},
		{name: "width", format: "%4w", numArgs: 1, want: []int{0}},
		{name: "width and precision", format: "%4.2w", numArgs: 1, want: []int{0}},
		{name: "star width", format: "%*w", numArgs: 2, want: []int{1}},
		{name: "star precision", format: "%.*w", numArgs: 2, want: []int{1}},
		{name: "star width and precision", format: "%*.*w", numArgs: 3, want: []int{2}},
		{name: "star width missing operand", format: "%*w", numArgs: 1, want: nil},
		{name: "star percent", format: "%*%%w", numArgs: 2, want: []int{1}},
		{name: "explicit index", format: "%[2]w", numArgs: 2, want: []int{1}},
		{name: "index repositions counter", format: "%[1]v %w", numArgs: 2, want: []int{1}},
		{name: "index continues", format: "%[2]w %w", numArgs: 3, want: []int{1, 2}},
		{name: "index on star width", format: "%[2]*w", numArgs: 3, want: []int{2}},
		{name: "multiple", format: "%w %w", numArgs: 2, want: []int{0, 1}},
		{name: "duplicate", format: "%[1]w %[1]w", numArgs: 1, want: []int{0, 0}},
		{name: "reordered", format: "%[2]w %[1]w", numArgs: 2, want: []int{1, 0}},
		{name: "index out of range", format: "%[3]w", numArgs: 2, want: nil},
		{name: "index out of range then wrap", format: "%[3]v %w", numArgs: 2, want: []int{0}},
		{name: "index then missing operand", format: "%[2]v %w", numArgs: 2, want: nil},
		{name: "index zero", format: "%[0]w", numArgs: 1, want: nil},
		{name: "empty index", format: "%[]w", numArgs: 1, want: nil},
		{name: "unclosed index", format: "%[1w", numArgs: 1, want: nil},
		{name: "width after index", format: "%[1]4w", numArgs: 1, want: nil},
		{name: "precision after index", format: "%[1].4w", numArgs: 1, want: nil},
		{name: "missing operand", format: "%v %w", numArgs: 1, want: nil},
		{name: "no operands", format: "%w", numArgs: 0, want: nil},
		{name: "extra operand", format: "%w", numArgs: 2, want: []int{0}},
		{name: "missing verb", format: "%", numArgs: 1, want: nil},
		{name: "trailing dot", format: "%.", numArgs: 1, want: nil},
		{name: "dot percent", format: "%.%w", numArgs: 1, want: nil},
		{name: "bad verb consumes", format: "%! %w", numArgs: 2, want: []int{1}},
		{name: "unicode verb", format: "%世 %w", numArgs: 2, want: []int{1}},
		{name: "unicode literal", format: "héllo %w", numArgs: 1, want: []int{0}},
		{name: "huge width", format: "%1000000000w", numArgs: 1, want: nil},
		{name: "huge precision", format: "%.1000000000w", numArgs: 1, want: nil},
		{name: "huge index", format: "%[1000000000]w", numArgs: 1, want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := slices.Collect(AllWrappedArgs(tt.format, tt.numArgs))
			if !slices.Equal(got, tt.want) {
				t.Errorf("AllWrappedArgs(%q, %d) = %v, want %v", tt.format, tt.numArgs, got, tt.want)
			}

			// Cross-check against the actual behavior of fmt.Errorf, which
			// returns the wrapped operands sorted by index.
			slices.Sort(got)
			got = slices.Compact(got)

			if want := fmtWrappedArgs(tt.format, tt.numArgs); !slices.Equal(got, want) {
				t.Errorf("AllWrappedArgs(%q, %d) = %v, but fmt.Errorf wraps %v", tt.format, tt.numArgs, got, want)
			}
		})
	}
}

// fmtWrappedArgs determines the indices of the operands fmt.Errorf actually
// wraps by unwrapping the error it returns, serving as a reference for
// wrappedArgs. The indices are sorted and deduplicated.
func fmtWrappedArgs(format string, numArgs int) []int {
	errs := make([]error, numArgs)
	args := make([]any, numArgs)

	for i := range numArgs {
		errs[i] = errors.New(strconv.Itoa(i))
		args[i] = errs[i]
	}

	var wrapped []error

	switch x := fmt.Errorf(format, args...).(type) {
	case interface{ Unwrap() []error }:
		wrapped = x.Unwrap()

	case interface{ Unwrap() error }:
		wrapped = []error{x.Unwrap()}
	}

	var indices []int

	for _, w := range wrapped {
		if i := slices.Index(errs, w); i >= 0 {
			indices = append(indices, i)
		}
	}

	// fmt.Errorf does this, but we don't depend on it.
	slices.Sort(indices)
	indices = slices.Compact(indices)

	return indices
}
