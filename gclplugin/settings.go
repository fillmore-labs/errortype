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

package gclplugin

import (
	errortype "fillmore-labs.com/errortype/analyzer"
	"fillmore-labs.com/errortype/detect"
)

// Settings are the linter settings.
type Settings struct {
	Overrides       Overrides `json:"overrides,omitzero"`
	StyleCheck      *bool     `json:"style-check,omitzero"`
	DeepIsCheck     *bool     `json:"deep-is-check,omitzero"`
	CheckIs         *bool     `json:"check-is,omitzero"`
	UncheckedAssert *bool     `json:"unchecked-assert,omitzero"`
	CheckUnused     *bool     `json:"check-unused,omitzero"`
}

// Overrides defines overrides for error types.
type Overrides struct {
	// Types that should be treated as pointer errors.
	Pointer []string `json:"pointer,omitempty"`
	// Types that should be treated as value errors.
	Value []string `json:"value,omitempty"`
	// Types for which error type checks should be suppressed.
	Suppress []string `json:"suppress,omitempty"`
}

// detectOptions converts Settings into [detect.Options] for the detect analyzer.
// It processes override configurations and returns them in the appropriate format.
func detectOptions(settings Settings) detect.Options {
	overrides := settings.Overrides

	var d detectOverrides
	d.setOverrides(detect.OverridePointer, overrides.Pointer)
	d.setOverrides(detect.OverrideValue, overrides.Value)
	d.setOverrides(detect.OverrideSuppress, overrides.Suppress)

	if len(d) == 0 {
		return nil
	}

	return detect.Options{detect.WithOverrides(d)}
}

type detectOverrides map[detect.Override][]string

func (d *detectOverrides) setOverrides(key detect.Override, values []string) {
	if len(values) > 0 {
		if *d == nil {
			*d = make(detectOverrides, 3)
		}

		(*d)[key] = values
	}
}

// Options converts [Settings] into a list of [errortype.Option] for the errortype analyzer.
// It processes settings and applies them only when explicitly set (non-nil).
func (s Settings) Options() []errortype.Option {
	var opts []errortype.Option

	opts = appendOption(opts, s.StyleCheck, errortype.WithStyleCheck)
	opts = appendOption(opts, s.DeepIsCheck, errortype.WithDeepIsCheck)
	opts = appendOption(opts, s.CheckIs, errortype.WithCheckIs)
	opts = appendOption(opts, s.UncheckedAssert, errortype.WithUncheckedAssert)
	opts = appendOption(opts, s.CheckUnused, errortype.WithCheckUnused)

	return opts
}

// appendOption appends a non-nil setting to an option list.
func appendOption[T, O any](opts []O, value *T, constructor func(T) O) []O {
	if value == nil {
		return opts
	}

	return append(opts, constructor(*value))
}
