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
	"cmp"
	"encoding"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"regexp"
	"slices"
	"strings"

	"fillmore-labs.com/errortype/detect/result"
	"fillmore-labs.com/errortype/internal/detect"
	"fillmore-labs.com/errortype/internal/options"
)

// Option configures specific behavior of the detect [analysis.Analyzer].
type Option interface {
	Apply(opts *detect.Options) error
	LogAttr() slog.Attr
}

// Join creates a new Option joining the provided Option values.
//
// The result implements [slog.LogValuer], so the following is evaluated lazily:
//
//	slog.LogAttrs(ctx, slog.LevelInfo, "settings", Join(opts...).LogAttr())
func Join(opts ...Option) Option {
	return options.Join(opts)
}

// WithOverrides returns an Option that applies the provided overrides mapping,
// allowing specific type names to be associated with custom error types.
func WithOverrides(overrides map[result.ErrorType][]string) Option {
	return overridesOption{overrides: overrides}
}

// WithWrappers returns an Option that applies the provided overrides mapping,
// allowing specific functions to be treated as a wrapper for standard function.
func WithWrappers(wrappers map[result.WrapperType][]string) Option {
	return wrappersOption{wrappers: wrappers}
}

// WithOverrideFile returns an [Option] that configures usage overrides by reading types from the specified file.
func WithOverrideFile(file string) Option {
	return overrideFileOption{file: file}
}

// WithHeuristics is an [Option] to configure heuristic passes.
func WithHeuristics(heuristics ...Heuristic) Option {
	return heuristicsOption{heuristics: heuristics}
}

// WithTrace is an [Option] to configure result output.
func WithTrace(trace *regexp.Regexp) Option { return traceOption{trace: trace} }

type overridesOption struct {
	overrides map[result.ErrorType][]string
}

func (o overridesOption) Apply(opts *detect.Options) error {
	return unmarshalAndApply(o.overrides, opts.AddOverrides)
}

func (o overridesOption) LogAttr() slog.Attr {
	return logAttrMap("overrides", o.overrides)
}

type wrappersOption struct {
	wrappers map[result.WrapperType][]string
}

func (o wrappersOption) Apply(opts *detect.Options) error {
	return unmarshalAndApply(o.wrappers, opts.AddWrappers)
}

func (o wrappersOption) LogAttr() slog.Attr {
	return logAttrMap("wrappers", o.wrappers)
}

type overrideFileOption struct {
	file string
}

func (o overrideFileOption) Apply(opts *detect.Options) error {
	return opts.ReadOverrides(o.file)
}

func (o overrideFileOption) LogAttr() slog.Attr {
	return slog.Any("overrideFile", o.file)
}

type heuristicsOption struct{ heuristics []Heuristic }

func (o heuristicsOption) Apply(opts *detect.Options) error {
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

		default:
			return fmt.Errorf("unknown heurisitc %s", heuristic)
		}
	}

	opts.Heuristics = combined

	return nil
}

func (o heuristicsOption) LogAttr() slog.Attr {
	heuristics := make([]string, 0, len(o.heuristics))
	for _, h := range o.heuristics {
		heuristics = append(heuristics, h.String())
	}

	return slog.String("heuristics", strings.Join(heuristics, ","))
}

type traceOption struct{ trace *regexp.Regexp }

func (o traceOption) Apply(opts *detect.Options) error { opts.Trace = o.trace; return nil }

func (o traceOption) LogAttr() slog.Attr {
	return slog.Any("trace", o.trace)
}

func logAttrMap[K interface {
	cmp.Ordered
	fmt.Stringer
}](name string, m map[K][]string) slog.Attr {
	keys := slices.Sorted(maps.Keys(m))
	attrs := make([]slog.Attr, 0, len(keys))

	for _, key := range keys {
		values := m[key]
		if len(values) == 0 {
			continue
		}

		attrs = append(attrs, slog.String(key.String(), strings.Join(values, ",")))
	}

	return slog.GroupAttrs(name, attrs...)
}

func unmarshalAndApply[V any, K comparable, PV interface {
	*V
	encoding.TextUnmarshaler
}](m map[K][]string, apply func(map[K][]V)) error {
	var errs []error

	res := make(map[K][]V, len(m))
	for k, names := range m {
		vs := make([]V, 0, len(names))
		for _, name := range names {
			i := len(vs)
			vs = vs[:i+1] // safe because cap(vs) == len(names)

			var pv PV = &vs[i]
			if err := pv.UnmarshalText([]byte(name)); err != nil {
				errs = append(errs, err)
				vs = vs[:i]
			}
		}

		res[k] = vs
	}

	if err := errors.Join(errs...); err != nil {
		return err
	}

	apply(res)

	return nil
}
