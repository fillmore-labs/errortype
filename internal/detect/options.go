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
	"log/slog"

	"fillmore-labs.com/errortype/internal/errortypes"
	"fillmore-labs.com/errortype/internal/overrides"
)

// HeuristicPass represents a set of heuristic flags used to control various passes in the analysis process.
type HeuristicPass uint8

const (
	// HeuristicUsage represents a heuristic pass for general usage.
	HeuristicUsage HeuristicPass = 1 << iota

	// HeuristicReceivers represents a heuristic pass for consistent method receivers.
	HeuristicReceivers
)

type options struct {
	// usageOverrides stores the usage configuration for error types, read from a file.
	usageOverrides map[string]map[string]errortypes.ErrorType

	// heuristics controls heuristic passes
	heuristics HeuristicPass

	// debug controls debug output
	debug bool
}

// defaultOptions returns a [options] struct initialized with default values.
func defaultOptions() *options {
	return &options{ // Default options
		usageOverrides: nil,
		heuristics:     HeuristicUsage | HeuristicReceivers,
		debug:          false,
	}
}

// makeOptions returns a [options] struct with overriding [Option]s applied.
func makeOptions(opts Options) *options {
	o := defaultOptions()
	opts.apply(o)

	return o
}

// Option configures specific behavior of the zerolint [analysis.Analyzer].
type Option interface {
	LogValue() slog.Value
	key() string
	apply(opts *options)
}

// Options is a list of [Option] values that also satisfies the [Option] interface.
type Options []Option

// LogValue implements the [slog.LogValuer] interface.
func (o Options) LogValue() slog.Value {
	as := make([]slog.Attr, 0, len(o))
	for _, opt := range o {
		as = append(as, slog.Attr{Key: opt.key(), Value: opt.LogValue()})
	}

	return slog.GroupValue(as...)
}

func (o Options) key() string { return "options" }

func (o Options) apply(opts *options) {
	for _, opt := range o {
		opt.apply(opts)
	}
}

// WithOverrides returns an Option that applies the provided overrides mapping,
// allowing specific type names to be associated with custom error types.
// The override map keys are type names, and the values are the corresponding error types
// to use for those types during error detection.
func WithOverrides(overrides []overrides.Override) Option {
	return overridesOption{overrides: overrides}
}

type overridesOption struct {
	overrides []overrides.Override
}

// LogValue implements Option.
func (o overridesOption) LogValue() slog.Value {
	var as []slog.Attr
	for _, usage := range o.overrides {
		as = append(as, slog.Attr{
			Key:   usage.TypeName.String(),
			Value: slog.StringValue(usage.ErrorType.String()),
		})
	}

	return slog.GroupValue(as...)
}

func (o overridesOption) key() string { return "overrides" }

func (o overridesOption) apply(opts *options) { opts.addOverrides(o.overrides) }

// WithHeuristics is an [Option] to configure heuristic passes.
func WithHeuristics(heuristics ...HeuristicPass) Option {
	var combined HeuristicPass
	for _, heuristic := range heuristics {
		combined |= heuristic
	}

	return heuristicsOption{heuristics: combined}
}

type heuristicsOption struct{ heuristics HeuristicPass }

// LogValue implements the [slog.LogValuer] interface.
func (o heuristicsOption) LogValue() slog.Value {
	var v []string
	if o.heuristics == 0 {
		v = append(v, "None")
	} else {
		for _, mask := range [...]struct {
			name      string
			heuristic HeuristicPass
		}{
			{"usage", HeuristicUsage},
			{"receivers", HeuristicReceivers},
		} {
			if o.heuristics&mask.heuristic != 0 {
				v = append(v, mask.name)
			}
		}
	}

	return slog.AnyValue(v)
}

func (o heuristicsOption) key() string { return "heuristics" }

func (o heuristicsOption) apply(opts *options) { opts.heuristics = o.heuristics }

// WithDebug is an [Option] to configure debug output.
func WithDebug(debug bool) Option { return debugOption{debug: debug} }

type debugOption struct{ debug bool }

// LogValue implements the [slog.LogValuer] interface.
func (o debugOption) LogValue() slog.Value { return slog.BoolValue(o.debug) }

func (o debugOption) key() string { return "debug" }

func (o debugOption) apply(opts *options) { opts.debug = o.debug }
