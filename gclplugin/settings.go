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

import errortype "fillmore-labs.com/errortype/analyzer"

// Settings are the linter settings.
type Settings struct {
	Recommended     *bool     `json:"recommended,omitzero"`
	NonComparable   *bool     `json:"non-comparable,omitzero"`
	Naming          *bool     `json:"naming,omitzero"`
	Legacy          *bool     `json:"legacy,omitzero"`
	Receivers       *bool     `json:"receivers,omitzero"`
	StyleCheck      *bool     `json:"style-check,omitzero"`
	StrictIsTarget  *bool     `json:"strict-is-target,omitzero"`
	StrictIsMethod  *bool     `json:"strict-is-method,omitzero"`
	UncheckedAssert *bool     `json:"unchecked-assert,omitzero"`
	CheckUnused     *bool     `json:"check-unused,omitzero"`
	PrefixFilter    *bool     `json:"prefix-filter,omitzero"`
	Wrappers        Wrappers  `json:"wrappers,omitzero"`
	Overrides       Overrides `json:"overrides,omitzero"`
}

// Overrides defines overrides for error types, given as fully qualified type names.
type Overrides struct {
	// Types that should be treated as pointer errors.
	Pointer []string `json:"pointer,omitempty"`

	// Types that should be treated as value errors.
	Value []string `json:"value,omitempty"`

	// Types for which error type checks should be suppressed.
	Suppress []string `json:"suppress,omitempty"`
}

// Wrappers override wrapper function autodetection in packages, see [errortype.WithWrappers].
type Wrappers map[string][]string

// options converts [Settings] into a list of [errortype.Option] for the errortype analyzer.
// It processes settings and applies them only when explicitly set (non-nil).
func (s Settings) options() ([]errortype.Option, error) {
	var opts []errortype.Option

	opts = appendOption(opts, s.Recommended, errortype.WithRecommended)
	opts = appendOption(opts, s.NonComparable, errortype.WithNonComparable)
	opts = appendOption(opts, s.Naming, errortype.WithNaming)
	opts = appendOption(opts, s.Legacy, errortype.WithLegacy)
	opts = appendOption(opts, s.Receivers, errortype.WithReceivers)
	opts = appendOption(opts, s.StyleCheck, errortype.WithStyleCheck)
	opts = appendOption(opts, s.StrictIsTarget, errortype.WithStrictIsTarget)
	opts = appendOption(opts, s.StrictIsMethod, errortype.WithStrictIsMethod)
	opts = appendOption(opts, s.UncheckedAssert, errortype.WithUncheckedAssert)
	opts = appendOption(opts, s.CheckUnused, errortype.WithCheckUnused)
	opts = appendOption(opts, s.PrefixFilter, errortype.WithPrefixFilter)

	if overrides := s.Overrides.toMap(); len(overrides) > 0 {
		opts = append(opts, errortype.WithOverrides(overrides))
	}

	if len(s.Wrappers) > 0 {
		opts = append(opts, errortype.WithWrappers(s.Wrappers))
	}

	return opts, nil
}

// appendOption appends a non-nil setting to an option list.
func appendOption[T, O any](opts []O, value *T, constructor func(T) O) []O {
	if value == nil {
		return opts
	}

	return append(opts, constructor(*value))
}

// toMap converts [Overrides] into the form expected by [errortype.WithOverrides].
func (o Overrides) toMap() map[string][]string {
	overrides := make(map[string][]string, 3)

	for kind, names := range map[string][]string{
		"pointer":  o.Pointer,
		"value":    o.Value,
		"suppress": o.Suppress,
	} {
		if len(names) > 0 {
			overrides[kind] = names
		}
	}

	return overrides
}
