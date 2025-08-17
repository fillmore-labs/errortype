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
	"os"
	"path/filepath"
	"strings"

	"fillmore-labs.com/errortype/internal/overrides"
)

// readOverrides reads error type usage overrides from the specified file.
// If fileName is empty, no action is taken.
func (o *options) readOverrides(fileName string) error {
	if fileName == "" {
		return nil
	}

	overridesFile, err := os.Open(filepath.Clean(fileName))
	if err != nil {
		return fmt.Errorf("can't open overrides file: %w", err)
	}

	defer overridesFile.Close()

	usageOverrides, err := overrides.Read(overridesFile)
	if err != nil {
		return fmt.Errorf("can't read overrides file %s: %w", fileName, err)
	}

	o.addOverrides(usageOverrides)

	return nil
}

// setHeuristics parses and sets the heuristic passes from a comma-separated list.
// Valid values are: "usage", "receivers", and "off".
// "off" disables all heuristics and cannot be combined with other values.
func (o *options) setHeuristics(list string) error {
	var (
		heuristics HeuristicPass
		hasOff     bool
	)

	for h := range splitComma(list) {
		switch h {
		case "":

		case HeuristicOff.String():
			hasOff = true

		case HeuristicUsage.String():
			heuristics |= HeuristicUsage

		case HeuristicReceivers.String():
			heuristics |= HeuristicReceivers

		default:
			return fmt.Errorf("unknown heuristic %q", h) //nolint:err113
		}
	}

	if hasOff && heuristics != 0 {
		return fmt.Errorf(`heuristic %q cannot be combined with other values in %q`, HeuristicOff.String(), list) //nolint:err113
	}

	// Only update if the user provided some values.
	if heuristics != 0 || hasOff {
		o.heuristics = heuristics
	}

	return nil
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
