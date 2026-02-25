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
	"encoding"
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

// WithNonComparable is an [Option] to configure extended checking for non-comparable error types.
func WithNonComparable(nonComparable bool) Option {
	return analyzeOption{mask: settings.OptionNonComparable, value: nonComparable}
}

// WithNaming is an [Option] to configure name checking for error types and sentinel values.
func WithNaming(naming bool) Option {
	return analyzeOption{mask: settings.OptionNaming, value: naming}
}

// WithLegacy is an [Option] to configure checking for legacy pre-Go 1.13 error assertion queries.
func WithLegacy(legacy bool) Option {
	return analyzeOption{mask: settings.OptionLegacy, value: legacy}
}

// WithReceivers is an [Option] to configure checking error type receivers: a pointer receiver on Error
// for pointer errors, and value receivers on exported methods for value errors, except decoders like
// UnmarshalJSON.
func WithReceivers(receivers bool) Option {
	return analyzeOption{mask: settings.OptionReceivers, value: receivers}
}

// WithStyleCheck is an [Option] to configure the style check.
func WithStyleCheck(styleCheck bool) Option {
	return analyzeOption{mask: settings.OptionStyleCheck, value: styleCheck}
}

// WithStrictIsTarget is an [Option] to report “errors.Is” calls with a new address
// even when an “Is(error) bool” or “Unwrap() error” method could match it.
func WithStrictIsTarget(strictIsTarget bool) Option {
	return analyzeOption{mask: settings.OptionStrictIsTarget, value: strictIsTarget}
}

// WithStrictIsMethod is an [Option] to report all unwrapping calls in “Is” methods,
// not only those on the target.
func WithStrictIsMethod(strictIsMethod bool) Option {
	return analyzeOption{mask: settings.OptionStrictIsMethod, value: strictIsMethod}
}

// WithUncheckedAssert is an [Option] to configure diagnosis of an unchecked type assert on errors.
func WithUncheckedAssert(uncheckedAssert bool) Option {
	return analyzeOption{mask: settings.OptionUncheckedAssert, value: uncheckedAssert}
}

// WithCheckUnused is an [Option] to configure diagnosis of unused results of “errors.Is/As/AsType”, “fmt.Errorf” and their wrappers.
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

func logAttrMap(name string, m map[string][]string) slog.Attr {
	keys := slices.Sorted(maps.Keys(m))
	attrs := make([]slog.Attr, 0, len(keys))

	for _, key := range keys {
		values := m[key]
		if len(values) == 0 {
			continue
		}

		attrs = append(attrs, slog.String(key, strings.Join(values, ",")))
	}

	return slog.GroupAttrs(name, attrs...)
}

// parseNames parses a map of names in their text form, as used in override files,
// in key order, so the first error reported is deterministic.
func parseNames[K comparable, V any, PV interface {
	*V
	encoding.TextUnmarshaler
}](m map[string][]string, parseKey func(string) (K, error)) (map[K][]V, error) {
	if len(m) == 0 {
		return nil, nil
	}

	parsed := make(map[K][]V, len(m))

	for _, key := range slices.Sorted(maps.Keys(m)) {
		k, err := parseKey(key)
		if err != nil {
			return nil, err
		}

		for _, name := range m[key] {
			var v V
			if err := PV(&v).UnmarshalText([]byte(name)); err != nil {
				return nil, fmt.Errorf("%s: %w", key, err)
			}

			parsed[k] = append(parsed[k], v)
		}
	}

	return parsed, nil
}
