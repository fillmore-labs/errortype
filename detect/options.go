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
	"regexp"
	"slices"
	"strings"

	"fillmore-labs.com/errortype/facts"
	"fillmore-labs.com/errortype/internal/detect"
	"fillmore-labs.com/errortype/internal/overrides"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// Option configures specific behavior of the detect [analysis.Analyzer].
type Option interface {
	apply(opts *detect.Options)
	LogAttr() slog.Attr
}

// Options is a list of [Option] values that also satisfies the [Option] interface.
type Options []Option

func (o Options) apply(opts *detect.Options) {
	for _, opt := range o {
		opt.apply(opts)
	}
}

// LogValue implements [slog.LogValuer].
func (o Options) LogValue() slog.Value {
	as := make([]slog.Attr, 0, len(o))
	for _, opt := range o {
		as = append(as, opt.LogAttr())
	}

	return slog.GroupValue(as...)
}

// LogAttr returns a [slog.Attr] for logging.
func (o Options) LogAttr() slog.Attr {
	return slog.Any("options", o)
}

// WithOverrides returns an Option that applies the provided overrides mapping,
// allowing specific type names to be associated with custom error types.
func WithOverrides(overrides map[Override][]string) Option {
	return overridesOption{overrides: overrides}
}

type overridesOption struct {
	overrides map[Override][]string
}

func (o overridesOption) apply(opts *detect.Options) {
	or := make(overrides.Overrides)

	for typ, names := range o.overrides {
		var et facts.ErrorFact

		switch typ {
		case OverridePointer:
			et = facts.PointerType

		case OverrideValue:
			et = facts.ValueType

		case OverrideSuppress:
			et = facts.SuppressType

		default:
			continue
		}

		l := slices.Grow(or[et], len(names))
		for _, name := range names {
			var tn typeutil.TypeName
			if err := tn.UnmarshalText([]byte(name)); err != nil {
				continue
			}

			l = append(l, tn)
		}
		or[et] = l
	}

	opts.AddOverrides(or)
}

func (o overridesOption) LogAttr() slog.Attr {
	var as []slog.Attr
	for override, usage := range o.overrides {
		as = append(as, slog.Attr{
			Key:   override.String(),
			Value: slog.StringValue(strings.Join(usage, ",")),
		})
	}

	// go1.25: return slog.GroupAttrs("overrides", as...)
	return slog.Attr{Key: "overrides", Value: slog.GroupValue(as...)}
}

// WithOverrideFile returns an [Option] that configures usage overrides by reading types from the specified file.
func WithOverrideFile(file string) Option {
	return overrideFileOption{file: file}
}

type overrideFileOption struct {
	file string
}

func (o overrideFileOption) apply(opts *detect.Options) {
	if err := opts.ReadOverrides(o.file); err != nil {
		opts.InitializationError = err
	}
}

func (o overrideFileOption) LogAttr() slog.Attr {
	return slog.Any("overrideFile", o.file)
}

// WithHeuristics is an [Option] to configure heuristic passes.
func WithHeuristics(heuristics ...Heuristic) Option {
	return heuristicsOption{heuristics: heuristics}
}

type heuristicsOption struct{ heuristics []Heuristic }

func (o heuristicsOption) apply(opts *detect.Options) {
	var combined detect.Heuristics

	for _, heuristic := range o.heuristics {
		switch heuristic {
		case HeuristicOff:
			combined = detect.HeuristicOff

		case HeuristicVar:
			combined |= detect.HeuristicVar

		case HeuristicUsage:
			combined |= detect.HeuristicUsage

		case HeuristicReceivers:
			combined |= detect.HeuristicReceivers
		}
	}

	opts.Heuristics = combined
}

func (o heuristicsOption) LogAttr() slog.Attr {
	heuristics := make([]string, 0, len(o.heuristics))
	for _, h := range o.heuristics {
		heuristics = append(heuristics, h.String())
	}

	return slog.String("heuristics", strings.Join(heuristics, ","))
}

// WithTrace is an [Option] to configure result output.
func WithTrace(trace *regexp.Regexp) Option { return traceOption{trace: trace} }

type traceOption struct{ trace *regexp.Regexp }

func (o traceOption) apply(opts *detect.Options) { opts.Trace = o.trace }

func (o traceOption) LogAttr() slog.Attr {
	var re string
	if o.trace != nil {
		re = o.trace.String()
	}

	return slog.String("trace", re)
}
