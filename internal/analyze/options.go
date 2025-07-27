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

package analyze

import (
	"log/slog"

	"golang.org/x/tools/go/analysis"
)

type options struct {
	detecttypes *analysis.Analyzer

	// styleCheck controls style check
	styleCheck bool
}

// defaultOptions returns a [options] struct initialized with default values.
func defaultOptions() *options {
	return &options{ // Default options
		detecttypes: nil,
		styleCheck:  true,
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

func (o Options) apply(opts *options) {
	for _, opt := range o {
		opt.apply(opts)
	}
}

func (o Options) key() string {
	return "options"
}

// WithDetectTypes sets a custom *analysis.Analyzer for detecting error types.
func WithDetectTypes(detecttypes *analysis.Analyzer) Option {
	return detectTypesOption{detecttypes: detecttypes}
}

type detectTypesOption struct{ detecttypes *analysis.Analyzer }

// LogValue implements the [slog.LogValuer] interface.
func (o detectTypesOption) LogValue() slog.Value { return slog.StringValue(o.detecttypes.Name) }

func (o detectTypesOption) key() string { return "detect" }

func (o detectTypesOption) apply(opts *options) { opts.detecttypes = o.detecttypes }

// WithStyleCheck is an [Option] to configure style check.
func WithStyleCheck(styleCheck bool) Option { return styleCheckOption{styleCheck: styleCheck} }

type styleCheckOption struct{ styleCheck bool }

// LogValue implements the [slog.LogValuer] interface.
func (o styleCheckOption) LogValue() slog.Value { return slog.BoolValue(o.styleCheck) }

func (o styleCheckOption) key() string { return "stylecheck" }

func (o styleCheckOption) apply(opts *options) { opts.styleCheck = o.styleCheck }
