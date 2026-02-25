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
	"log/slog"

	"fillmore-labs.com/errortype/detect"
	"fillmore-labs.com/errortype/internal/analyze"
	"fillmore-labs.com/errortype/internal/options"
	"fillmore-labs.com/errortype/internal/run"
)

// Option configures specific behavior of a [New] errortype [analysis.Analyzer].
type Option interface {
	Apply(opts *run.Options) error
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

// WithDetectOptions sets [detect.Option]s for detecting error types.
func WithDetectOptions(opts ...detect.Option) Option {
	return detectOption{opt: detect.Join(opts...)}
}

// WithGenerated is an [Option] to configure checking in generated files.
func WithGenerated(generated bool) Option {
	return analyzeOption{mask: analyze.OptionGenerated, value: generated}
}

// WithNotComparable is an [Option] to configure extended checking for not comparable error types.
func WithNotComparable(notcomparable bool) Option {
	return analyzeOption{mask: analyze.OptionNotComparable, value: notcomparable}
}

// WithNaming is an [Option] to configure name checking for error types and sentinel values.
func WithNaming(naming bool) Option {
	return analyzeOption{mask: analyze.OptionNaming, value: naming}
}

// WithLegacy is an [Option] to configure checking for legacy pre-Go 1.13 error assertion queries.
func WithLegacy(legacy bool) Option {
	return analyzeOption{mask: analyze.OptionLegacy, value: legacy}
}

// WithStyleCheck is an [Option] to configure the style check.
func WithStyleCheck(styleCheck bool) Option {
	return analyzeOption{mask: analyze.OptionStyleCheck, value: styleCheck}
}

// WithCheckIs is an [Option] that configures the diagnostic suppression behavior
// related to an “Is(error) bool” method.
func WithCheckIs(checkIs bool) Option {
	return analyzeOption{mask: analyze.OptionCheckIs, value: checkIs}
}

// WithDeepIsCheck is an [Option] to configure “Is” method analysis.
func WithDeepIsCheck(deepIsCheck bool) Option {
	return analyzeOption{mask: analyze.OptionDeepIsCheck, value: deepIsCheck}
}

// WithUncheckedAssert is an [Option] to configure diagnosis of an unchecked type assert on errors.
func WithUncheckedAssert(uncheckedAssert bool) Option {
	return analyzeOption{mask: analyze.OptionUncheckedAssert, value: uncheckedAssert}
}

// WithCheckUnused is an [Option] to configure diagnosis of unchecked results of “errors.As” calls.
func WithCheckUnused(checkUnused bool) Option {
	return analyzeOption{mask: analyze.OptionCheckUnused, value: checkUnused}
}

// WithPrefixFilter is an [Option] to configure prefix filtering for variable declarations.
func WithPrefixFilter(prefixFilter bool) Option {
	return analyzeOption{mask: analyze.OptionPrefixFilter, value: prefixFilter}
}

// WithRecommended is an [Option] to set recommended options.
func WithRecommended(value bool) Option {
	return recommendedOption{value: value}
}

type detectOption struct{ opt detect.Option }

func (o detectOption) Apply(opts *run.Options) error {
	return o.opt.Apply(opts.DetectOptions)
}

func (o detectOption) LogAttr() slog.Attr {
	return slog.Any("detect-options", o.opt)
}

type analyzeOption struct {
	mask  analyze.Options
	value bool
}

func (o analyzeOption) Apply(opts *run.Options) error {
	opts.Options.Set(o.mask, o.value)
	return nil
}

func (o analyzeOption) LogAttr() slog.Attr {
	return slog.Bool(o.mask.String(), o.value)
}

type recommendedOption struct {
	value bool
}

func (o recommendedOption) Apply(opts *run.Options) error {
	if o.value {
		opts.Options.Set(analyze.RecommendedOptions, true)
	}

	return nil
}

func (o recommendedOption) LogAttr() slog.Attr {
	return slog.Bool("recommended", o.value)
}
