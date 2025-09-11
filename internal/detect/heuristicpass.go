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
	"iter"
	"strings"
)

// HeuristicPass represents a set of heuristic flags used to control various passes in the analysis process.
type HeuristicPass uint8

const (
	// HeuristicUsage represents a heuristic pass for general usage.
	HeuristicUsage HeuristicPass = 1 << iota

	// HeuristicReceivers represents a heuristic pass for consistent method receivers.
	HeuristicReceivers

	// HeuristicOff turns off all heuristic passes.
	HeuristicOff HeuristicPass = 0
)

var heuristicPasses = map[HeuristicPass]string{
	HeuristicOff:       "off",
	HeuristicUsage:     "usage",
	HeuristicReceivers: "receivers",
}

// String returns the string representation of HeuristicPass.
// If multiple flags are set, it returns a comma-separated list of names.
// If no flags are set, it returns "off".
func (h HeuristicPass) String() string {
	if h == HeuristicOff {
		return heuristicPasses[HeuristicOff]
	}

	var parts []string

	for flag := HeuristicPass(1); flag != 0; flag <<= 1 {
		if h&flag != 0 {
			name, ok := heuristicPasses[flag]
			if !ok {
				name = fmt.Sprintf("Unknown(%d)", int(flag))
			}

			parts = append(parts, name)
		}
	}

	return strings.Join(parts, ", ")
}

// HeuristicsFromString parses a comma-separated string into a HeuristicPass value.
// Returns an error if the input contains invalid or conflicting heuristics.
func HeuristicsFromString(list string) (HeuristicPass, error) {
	var (
		heuristics HeuristicPass
		hasOff     bool
	)

	for h := range splitComma(list) {
		switch h {
		case "":

		case heuristicPasses[HeuristicOff]:
			hasOff = true

		case heuristicPasses[HeuristicUsage]:
			heuristics |= HeuristicUsage

		case heuristicPasses[HeuristicReceivers]:
			heuristics |= HeuristicReceivers

		default:
			return HeuristicOff, fmt.Errorf("unknown heuristic %q", h)
		}
	}

	if hasOff && heuristics != 0 {
		return HeuristicOff, fmt.Errorf(`heuristic %q cannot be combined with other values in %q`, HeuristicOff.String(), list)
	}

	return heuristics, nil
}

func splitComma(list string) iter.Seq[string] {
	fields := strings.Split(list, ",")

	return func(yield func(string) bool) {
		for _, s := range fields {
			trim := strings.TrimSpace(s)
			if trim == "" {
				continue
			}

			if !yield(trim) {
				return
			}
		}
	}
}
