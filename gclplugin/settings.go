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
	override := overrideOption(settings.Overrides)
	if len(override) == 0 {
		return nil
	}

	return detect.Options{detect.WithOverrides(override)}
}

// overrideOption transforms the [Overrides] struct into a map format expected by
// the detect package. It only includes overrides that have non-empty type lists.
func overrideOption(overrides Overrides) map[detect.Override][]string {
	overrideConfigs := [...]struct {
		override detect.Override
		types    []string
	}{
		{detect.OverridePointer, overrides.Pointer},
		{detect.OverrideValue, overrides.Value},
		{detect.OverrideSuppress, overrides.Suppress},
	}

	override := make(map[detect.Override][]string, 3)

	for _, opt := range overrideConfigs {
		if len(opt.types) > 0 {
			override[opt.override] = opt.types
		}
	}

	return override
}

// errortypeOptions converts [Settings] into [errortype.Options] for the errortype analyzer.
// It processes boolean settings and applies them only when explicitly set (non-nil).
func errortypeOptions(settings Settings) errortype.Options {
	optionConfigs := [...]struct {
		f func(bool) errortype.Option
		v *bool
	}{
		{errortype.WithStyleCheck, settings.StyleCheck},
		{errortype.WithDeepIsCheck, settings.DeepIsCheck},
		{errortype.WithCheckIs, settings.CheckIs},
		{errortype.WithUncheckedAssert, settings.UncheckedAssert},
		{errortype.WithCheckUnused, settings.CheckUnused},
	}

	var options errortype.Options

	for _, opt := range optionConfigs {
		if opt.v != nil {
			options = append(options, opt.f(*opt.v))
		}
	}

	return options
}
