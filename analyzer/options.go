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

	"golang.org/x/tools/go/analysis"

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

// WithDetectTypes sets a custom *[analysis.Analyzer] for detecting error types.
func WithDetectTypes(detectTypes *analysis.Analyzer) Option {
	return detectTypesOption{detectTypes: detectTypes}
}

// WithStyleCheck is an [Option] to configure the style check.
func WithStyleCheck(styleCheck bool) Option {
	return flagOption{flag: analyze.OptionStyleCheck, value: styleCheck}
}

// WithCheckIs is an [Option] that configures the diagnostic suppression behavior
// related to an “Is(error) bool” method.
func WithCheckIs(checkIs bool) Option {
	return flagOption{flag: analyze.OptionCheckIs, value: checkIs}
}

// WithDeepIsCheck is an [Option] to configure “Is” method analysis.
func WithDeepIsCheck(deepIsCheck bool) Option {
	return flagOption{flag: analyze.OptionDeepIsCheck, value: deepIsCheck}
}

// WithUncheckedAssert is an [Option] to configure diagnosis of an unchecked type assert on errors.
func WithUncheckedAssert(uncheckedAssert bool) Option {
	return flagOption{flag: analyze.OptionUncheckedAssert, value: uncheckedAssert}
}

// WithCheckUnused is an [Option] to configure diagnosis of unchecked results of “errors.As” calls.
func WithCheckUnused(checkUnused bool) Option {
	return flagOption{flag: analyze.OptionCheckUnused, value: checkUnused}
}

// WithNaming is an [Option] to configure name checking for error types and sentinel values.
func WithNaming(naming bool) Option {
	return flagOption{flag: analyze.OptionNaming, value: naming}
}

// WithPrefixFilter is an [Option] to configure prefix filtering for variable declarations.
func WithPrefixFilter(prefixFilter bool) Option {
	return flagOption{flag: analyze.OptionPrefixFilter, value: prefixFilter}
}

type detectTypesOption struct{ detectTypes *analysis.Analyzer }

func (o detectTypesOption) Apply(opts *run.Options) error {
	opts.DetectTypes = o.detectTypes
	return nil
}

func (o detectTypesOption) LogAttr() slog.Attr {
	return slog.String("detect", o.detectTypes.Name)
}

type flagOption struct {
	flag  analyze.Options
	value bool
}

func (o flagOption) Apply(opts *run.Options) error {
	opts.SetOption(o.flag, o.value)
	return nil
}

func (o flagOption) LogAttr() slog.Attr {
	return slog.Bool(o.flag.String(), o.value)
}
