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

package detect

import (
	"fmt"
	"strings"
	"unicode"

	"fillmore-labs.com/errortype/internal/bitflag"
)

// Heuristics represents a set of heuristic flags used to control various passes in the analysis process.
type Heuristics uint8

const (
	// HeuristicVar represents a heuristic pass for variable declarations.
	HeuristicVar Heuristics = 1 << posHeuristicVar

	// HeuristicUsage represents a heuristic pass for general usage.
	HeuristicUsage Heuristics = 1 << posHeuristicUsage

	// HeuristicReceivers represents a heuristic pass for consistent method receivers.
	HeuristicReceivers Heuristics = 1 << posHeuristicReceivers

	// HeuristicAll combines all available heuristic passes into a single constant, encompassing all analysis strategies.
	HeuristicAll = HeuristicVar | HeuristicUsage | HeuristicReceivers

	// HeuristicOff turns off all heuristic passes.
	HeuristicOff Heuristics = 0
)

const (
	posHeuristicVar = 2 - iota
	posHeuristicUsage
	posHeuristicReceivers
)

var _heuristicPasses = [...]string{
	posHeuristicVar:       "var",
	posHeuristicUsage:     "usage",
	posHeuristicReceivers: "receivers",
}

// String returns the string representation of HeuristicPass.
// If multiple flags are set, it returns a comma-separated list of names.
// If no flags are set, it returns "off".
func (h Heuristics) String() string {
	return bitflag.ToString(h, _heuristicPasses[:], "off")
}

// HeuristicsFromString parses a comma-separated string into a HeuristicPass value.
// Returns an error if the input contains invalid or conflicting heuristics.
func HeuristicsFromString(list string) (Heuristics, error) {
	var (
		heuristics Heuristics
		hasOff     bool
	)

	for h := range strings.FieldsFuncSeq(list, commaOrSpace) {
		switch h {
		case "":
			continue

		case "off":
			hasOff = true

		case _heuristicPasses[posHeuristicVar]:
			heuristics |= HeuristicVar

		case _heuristicPasses[posHeuristicUsage]:
			heuristics |= HeuristicUsage

		case _heuristicPasses[posHeuristicReceivers]:
			heuristics |= HeuristicReceivers

		default:
			return HeuristicOff, fmt.Errorf("unknown heuristic %q", h)
		}
	}

	if hasOff && heuristics != HeuristicOff {
		return HeuristicOff, fmt.Errorf(`heuristic %q cannot be combined with other values in %q`, HeuristicOff.String(), list)
	}

	return heuristics, nil
}

func commaOrSpace(r rune) bool {
	return r == ',' || unicode.IsSpace(r)
}
