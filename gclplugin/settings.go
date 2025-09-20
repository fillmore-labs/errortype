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

func detectOptions(settings Settings) []detect.Option {
	var opts []detect.Option

	overrides := make(map[detect.Override][]string)

	for _, opt := range []struct {
		types    []string
		override detect.Override
	}{
		{settings.Overrides.Pointer, detect.OverridePointer},
		{settings.Overrides.Value, detect.OverrideValue},
		{settings.Overrides.Suppress, detect.OverrideSuppress},
	} {
		if len(opt.types) > 0 {
			overrides[opt.override] = opt.types
		}
	}

	if len(overrides) > 0 {
		opts = append(opts, detect.WithOverrides(overrides))
	}

	return opts
}

func errortypeOptions(settings Settings) []errortype.Option {
	var opts []errortype.Option

	for _, opt := range []struct {
		v *bool
		f func(bool) errortype.Option
	}{
		{settings.StyleCheck, errortype.WithStyleCheck},
		{settings.DeepIsCheck, errortype.WithDeepIsCheck},
		{settings.CheckIs, errortype.WithCheckIs},
		{settings.UncheckedAssert, errortype.WithUncheckedAssert},
	} {
		if opt.v != nil {
			opts = append(opts, opt.f(*opt.v))
		}
	}

	return opts
}
