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
)

// makeOptions returns a [analyze.Options] struct with overriding [Options] applied.
func makeOptions(opts Options) *analyze.Options {
	o := analyze.DefaultOptions()
	opts.apply(o)

	return o
}

// Option configures specific behavior of a [New] errortype [analysis.Analyzer].
type Option interface {
	apply(opts *analyze.Options)
	SlogAttr() slog.Attr
}

// Options is a list of [Option] values that also satisfies the [Option] interface.
type Options []Option

func (o Options) apply(opts *analyze.Options) {
	for _, opt := range o {
		opt.apply(opts)
	}
}

// LogValue implements [slog.LogValuer].
func (o Options) LogValue() slog.Value {
	as := make([]slog.Attr, 0, len(o))
	for _, opt := range o {
		as = append(as, opt.SlogAttr())
	}

	return slog.GroupValue(as...)
}

// SlogAttr returns a [slog.Attr] for logging.
func (o Options) SlogAttr() slog.Attr {
	return slog.Any("options", o)
}

// WithDetectTypes sets a custom *[analysis.Analyzer] for detecting error types.
func WithDetectTypes(detecttypes *analysis.Analyzer) Option {
	return detectTypesOption{detecttypes: detecttypes}
}

type detectTypesOption struct{ detecttypes *analysis.Analyzer }

func (o detectTypesOption) apply(opts *analyze.Options) { opts.DetectTypes = o.detecttypes }

func (o detectTypesOption) SlogAttr() slog.Attr {
	return slog.String("detect", o.detecttypes.Name)
}

// WithStyleCheck is an [Option] to configure the style check.
func WithStyleCheck(styleCheck bool) Option { return styleCheckOption{styleCheck: styleCheck} }

type styleCheckOption struct{ styleCheck bool }

func (o styleCheckOption) apply(opts *analyze.Options) { opts.StyleCheck = o.styleCheck }

func (o styleCheckOption) SlogAttr() slog.Attr {
	return slog.Bool("styleCheck", o.styleCheck)
}

// WithCheckIs is an [Option] that configures the diagnostic suppression behavior
// related to the `Is(error) bool` method.
// If `checkIs` is true (the default), diagnostics for `errors.Is(err, &MyError{})`
// may be suppressed if `*MyError` implements `Is(error) bool`.
// If false, this specific suppression heuristic is disabled.
func WithCheckIs(checkIs bool) Option { return checkIsOption{checkIs: checkIs} }

type checkIsOption struct{ checkIs bool }

func (o checkIsOption) apply(opts *analyze.Options) { opts.CheckIs = o.checkIs }

func (o checkIsOption) SlogAttr() slog.Attr {
	return slog.Bool("checkIs", o.checkIs)
}

// WithDeepIsCheck is an [Option] to configure `Is` method analysis.
// If deepIsCheck is true, the analyzer will flag every method calling `Unwrap`.
// The default behavior is to flag only applications to `target“.
func WithDeepIsCheck(deepIsCheck bool) Option {
	return deepIsCheckOption{deepIsCheck: deepIsCheck}
}

type deepIsCheckOption struct{ deepIsCheck bool }

func (o deepIsCheckOption) apply(opts *analyze.Options) { opts.DeepIsCheck = o.deepIsCheck }

func (o deepIsCheckOption) SlogAttr() slog.Attr {
	return slog.Bool("deepIsCheck", o.deepIsCheck)
}

// WithUncheckedAssert is an [Option] to configure diagnosis of an unchecked type assert on errors.
func WithUncheckedAssert(uncheckedAssert bool) Option {
	return uncheckedAssertOption{uncheckedAssert: uncheckedAssert}
}

type uncheckedAssertOption struct{ uncheckedAssert bool }

func (o uncheckedAssertOption) apply(opts *analyze.Options) { opts.UncheckedAssert = o.uncheckedAssert }

func (o uncheckedAssertOption) SlogAttr() slog.Attr {
	return slog.Bool("uncheckedAssert", o.uncheckedAssert)
}
