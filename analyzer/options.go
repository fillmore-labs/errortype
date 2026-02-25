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

package analyzer

import (
	"cmp"
	"fmt"
	"log/slog"
	"maps"
	"regexp"
	"slices"
	"strings"

	"fillmore-labs.com/errortype/internal/options"
	"fillmore-labs.com/errortype/internal/settings"
)

// Option configures specific behavior of a [New] errortype [analysis.Analyzer].
type Option interface {
	Apply(opts *settings.Config) error
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

// WithGenerated is an [Option] to configure checking in generated files.
func WithGenerated(generated bool) Option {
	return analyzeOption{mask: settings.OptionGenerated, value: generated}
}

// WithNotComparable is an [Option] to configure extended checking for not comparable error types.
func WithNotComparable(notcomparable bool) Option {
	return analyzeOption{mask: settings.OptionNotComparable, value: notcomparable}
}

// WithNaming is an [Option] to configure name checking for error types and sentinel values.
func WithNaming(naming bool) Option {
	return analyzeOption{mask: settings.OptionNaming, value: naming}
}

// WithLegacy is an [Option] to configure checking for legacy pre-Go 1.13 error assertion queries.
func WithLegacy(legacy bool) Option {
	return analyzeOption{mask: settings.OptionLegacy, value: legacy}
}

// WithStyleCheck is an [Option] to configure the style check.
func WithStyleCheck(styleCheck bool) Option {
	return analyzeOption{mask: settings.OptionStyleCheck, value: styleCheck}
}

// WithCheckIs is an [Option] that configures the diagnostic suppression behavior
// related to an “Is(error) bool” method.
func WithCheckIs(checkIs bool) Option {
	return analyzeOption{mask: settings.OptionCheckIs, value: checkIs}
}

// WithDeepIsCheck is an [Option] to configure “Is” method analysis.
func WithDeepIsCheck(deepIsCheck bool) Option {
	return analyzeOption{mask: settings.OptionDeepIsCheck, value: deepIsCheck}
}

// WithUncheckedAssert is an [Option] to configure diagnosis of an unchecked type assert on errors.
func WithUncheckedAssert(uncheckedAssert bool) Option {
	return analyzeOption{mask: settings.OptionUncheckedAssert, value: uncheckedAssert}
}

// WithCheckUnused is an [Option] to configure diagnosis of unchecked results of “errors.As” calls.
func WithCheckUnused(checkUnused bool) Option {
	return analyzeOption{mask: settings.OptionCheckUnused, value: checkUnused}
}

// WithPrefixFilter is an [Option] to configure prefix filtering for variable declarations.
func WithPrefixFilter(prefixFilter bool) Option {
	return analyzeOption{mask: settings.OptionPrefixFilter, value: prefixFilter}
}

// WithRecommended is an [Option] to set recommended options.
func WithRecommended(value bool) Option {
	return recommendedOption{value: value}
}

// WithOverrideFile returns an [Option] that configures usage overrides by reading types from the specified file.
func WithOverrideFile(file string) Option {
	return overrideFileOption{file: file}
}

type analyzeOption struct {
	mask  settings.Options
	value bool
}

func (o analyzeOption) Apply(opts *settings.Config) error {
	opts.Options.Set(o.mask, o.value)
	return nil
}

func (o analyzeOption) LogAttr() slog.Attr {
	return slog.Bool(o.mask.String(), o.value)
}

type recommendedOption struct {
	value bool
}

func (o recommendedOption) Apply(opts *settings.Config) error {
	if o.value {
		opts.Options.Set(settings.RecommendedOptions, true)
	}

	return nil
}

func (o recommendedOption) LogAttr() slog.Attr {
	return slog.Bool("recommended", o.value)
}

type overrideFileOption struct {
	file string
}

func (o overrideFileOption) Apply(opts *settings.Config) error {
	return opts.ReadOverrides(o.file)
}

func (o overrideFileOption) LogAttr() slog.Attr {
	return slog.Any("override-file", o.file)
}

// WithTrace is an [Option] to configure result output.
func WithTrace(trace *regexp.Regexp) Option { return traceOption{trace: trace} }

type traceOption struct{ trace *regexp.Regexp }

func (o traceOption) Apply(opts *settings.Config) error {
	opts.TypeConfig.Trace = o.trace
	opts.WrapperConfig.Trace = o.trace

	return nil
}

func (o traceOption) LogAttr() slog.Attr {
	return slog.Any("trace", o.trace)
}

func logAttrMap[K interface {
	cmp.Ordered
	fmt.Stringer
}, V fmt.Stringer](name string, m map[K][]V) slog.Attr {
	keys := slices.Sorted(maps.Keys(m))
	attrs := make([]slog.Attr, 0, len(keys))

	for _, key := range keys {
		values := m[key]
		if len(values) == 0 {
			continue
		}

		vals := make([]string, len(values))
		for i, v := range values {
			vals[i] = v.String()
		}

		attrs = append(attrs, slog.String(key.String(), strings.Join(vals, ",")))
	}

	return slog.GroupAttrs(name, attrs...)
}
