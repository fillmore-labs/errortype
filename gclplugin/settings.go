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
	typesresult "fillmore-labs.com/errortype/internal/detecttypes/result"
	wrappersresult "fillmore-labs.com/errortype/internal/detectwrappers/result"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// Settings are the linter settings.
type Settings struct {
	Recommended     *bool     `json:"recommended,omitzero"`
	NotComparable   *bool     `json:"non-comparable,omitzero"`
	Naming          *bool     `json:"naming,omitzero"`
	Legacy          *bool     `json:"legacy,omitzero"`
	StyleCheck      *bool     `json:"style-check,omitzero"`
	DeepIsCheck     *bool     `json:"deep-is-check,omitzero"`
	CheckIs         *bool     `json:"check-is,omitzero"`
	UncheckedAssert *bool     `json:"unchecked-assert,omitzero"`
	CheckUnused     *bool     `json:"check-unused,omitzero"`
	PrefixFilter    *bool     `json:"prefix-filter,omitzero"`
	Wrappers        Wrappers  `json:"wrappers,omitzero"`
	Overrides       Overrides `json:"overrides,omitzero"`
}

// Overrides defines overrides for error types.
type Overrides struct {
	// Types that should be treated as pointer errors.
	Pointer []typeutil.TypeName `json:"pointer,omitempty"`

	// Types that should be treated as value errors.
	Value []typeutil.TypeName `json:"value,omitempty"`

	// Types for which error type checks should be suppressed.
	Suppress []typeutil.TypeName `json:"suppress,omitempty"`
}

// Wrappers override wrapper function autodetection in packages.
type Wrappers map[wrappersresult.WrapperType][]typeutil.FuncName

// options converts [Settings] into a list of [errortype.Option] for the errortype analyzer.
// It processes settings and applies them only when explicitly set (non-nil).
func (s Settings) options() ([]errortype.Option, error) {
	var opts []errortype.Option

	opts = appendOption(opts, s.Recommended, errortype.WithRecommended)
	opts = appendOption(opts, s.NotComparable, errortype.WithNotComparable)
	opts = appendOption(opts, s.Naming, errortype.WithNaming)
	opts = appendOption(opts, s.Legacy, errortype.WithLegacy)
	opts = appendOption(opts, s.StyleCheck, errortype.WithStyleCheck)
	opts = appendOption(opts, s.DeepIsCheck, errortype.WithDeepIsCheck)
	opts = appendOption(opts, s.CheckIs, errortype.WithCheckIs)
	opts = appendOption(opts, s.UncheckedAssert, errortype.WithUncheckedAssert)
	opts = appendOption(opts, s.CheckUnused, errortype.WithCheckUnused)
	opts = appendOption(opts, s.PrefixFilter, errortype.WithPrefixFilter)

	opts = appendOverridesOption(opts, s.Overrides)
	opts = appendWrappersOption(opts, s.Wrappers)

	return opts, nil
}

// appendOption appends a non-nil setting to an option list.
func appendOption[T, O any](opts []O, value *T, constructor func(T) O) []O {
	if value == nil {
		return opts
	}

	return append(opts, constructor(*value))
}

// appendOverridesOption converts [Overrides] into options for the detecttypes analyzer.
func appendOverridesOption(opts []errortype.Option, overrides Overrides) []errortype.Option {
	var or typeOverrides

	or.setOverrides(typesresult.Pointer, overrides.Pointer)
	or.setOverrides(typesresult.Value, overrides.Value)
	or.setOverrides(typesresult.Suppress, overrides.Suppress)

	if len(or) > 0 {
		opts = append(opts, errortype.WithOverrides(or))
	}

	return opts
}

type typeOverrides map[typesresult.ErrorType][]typeutil.TypeName

func (o *typeOverrides) setOverrides(key typesresult.ErrorType, values []typeutil.TypeName) {
	if len(values) > 0 {
		if *o == nil {
			*o = make(typeOverrides, 3)
		}

		(*o)[key] = values
	}
}

// appendWrappersOption converts [Wrappers] into options for the detectwrappers analyzer.
func appendWrappersOption(opts []errortype.Option, wrappers Wrappers) []errortype.Option {
	if len(wrappers) > 0 {
		opts = append(opts, errortype.WithWrappers(wrappers))
	}

	return opts
}
