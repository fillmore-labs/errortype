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

package analyzer

import (
	"fmt"
	"log/slog"
	"strings"

	"fillmore-labs.com/errortype/internal/detecttypes/result"
	"fillmore-labs.com/errortype/internal/settings"
	"fillmore-labs.com/errortype/internal/settings/errortype"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// WithOverrides returns an Option that applies the provided overrides mapping,
// allowing specific type names to be associated with custom error types.
func WithOverrides(overrides map[result.ErrorType][]typeutil.TypeName) Option {
	return overridesOption{overrides: overrides}
}

type overridesOption struct {
	overrides map[result.ErrorType][]typeutil.TypeName
}

func (o overridesOption) Apply(opts *settings.Config) error {
	opts.TypeConfig.AddOverrides(o.overrides)
	return nil
}

func (o overridesOption) LogAttr() slog.Attr {
	return logAttrMap("overrides", o.overrides)
}

// WithHeuristics is an [Option] to configure heuristic passes.
func WithHeuristics(heuristics ...Heuristic) Option {
	return heuristicsOption{heuristics: heuristics}
}

// Heuristic represent a heuristic flag used to control various passes in the analysis process.
type Heuristic = errortype.Heuristic

const (
	// HeuristicVar represents a heuristic pass for variable declarations.
	HeuristicVar = errortype.HeuristicVar

	// HeuristicUsage represents a heuristic pass for general usage.
	HeuristicUsage = errortype.HeuristicUsage

	// HeuristicReceivers represents a heuristic pass for consistent method receivers.
	HeuristicReceivers = errortype.HeuristicReceivers

	// HeuristicAll combines all available heuristic passes into a single constant, encompassing all analysis strategies.
	HeuristicAll = errortype.HeuristicAll

	// HeuristicOff turns off all heuristic passes.
	HeuristicOff = errortype.HeuristicOff
)

type heuristicsOption struct{ heuristics []Heuristic }

func (o heuristicsOption) Apply(opts *settings.Config) error {
	var combined errortype.Heuristic

	for _, heuristic := range o.heuristics {
		switch heuristic {
		case HeuristicOff:
			combined = errortype.HeuristicOff

		case HeuristicVar:
			combined |= errortype.HeuristicVar

		case HeuristicUsage:
			combined |= errortype.HeuristicUsage

		case HeuristicReceivers:
			combined |= errortype.HeuristicReceivers

		case HeuristicAll:
			combined |= errortype.HeuristicAll

		default:
			return fmt.Errorf("unknown heurisitc %s", heuristic)
		}
	}

	opts.TypeConfig.Heuristics = combined

	return nil
}

func (o heuristicsOption) LogAttr() slog.Attr {
	heuristics := make([]string, 0, len(o.heuristics))
	for _, h := range o.heuristics {
		heuristics = append(heuristics, h.String())
	}

	return slog.String("heuristics", strings.Join(heuristics, ","))
}
