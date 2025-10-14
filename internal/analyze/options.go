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

import "fillmore-labs.com/errortype/internal/bitflag"

// Options represent configuration flags to control the behavior of style and correctness checks for errors.
type Options uint8

const (
	// OptionCheckIs controls whether to check for `Is(error) bool` methods.
	OptionCheckIs Options = 1 << posOptionCheckIs
	// OptionCheckUnused flags unchecked results of `errors.As` calls.
	OptionCheckUnused = 1 << posOptionCheckUnused
	// OptionDeepIsCheck flags all unwrap methods in `Is` method checks, not only those on target.
	OptionDeepIsCheck = 1 << posOptionDeepIsCheck
	// OptionStyleCheck controls the target style check in `errors.As` calls.
	OptionStyleCheck = 1 << posOptionStyleCheck
	// OptionUncheckedAssert flags all unchecked asserts on errors.
	OptionUncheckedAssert = 1 << posOptionUncheckedAssert

	// DefaultOptions is the default configuration for error analysis.
	DefaultOptions = OptionCheckIs | OptionStyleCheck
)

var _options = [...]string{
	posOptionCheckIs:         "checkIs",
	posOptionCheckUnused:     "checkUnused",
	posOptionDeepIsCheck:     "deepIsCheck",
	posOptionStyleCheck:      "styleCheck",
	posOptionUncheckedAssert: "uncheckedAssert",
}

const (
	posOptionCheckIs = 4 - iota
	posOptionCheckUnused
	posOptionDeepIsCheck
	posOptionStyleCheck
	posOptionUncheckedAssert
)

func (o Options) String() string {
	return bitflag.ToString(o, _options[:], "none")
}

// CheckIs controls whether to check for `Is(error) bool` methods.
func (o Options) CheckIs() bool {
	return o&OptionCheckIs != 0
}

// CheckUnused flags unchecked results of `errors.As` calls.
func (o Options) CheckUnused() bool {
	return o&OptionCheckUnused != 0
}

// DeepIsCheck flags all unwrap methods in `Is` method checks, not only those on target.
func (o Options) DeepIsCheck() bool {
	return o&OptionDeepIsCheck != 0
}

// StyleCheck controls the target style check in `errors.As` calls.
func (o Options) StyleCheck() bool {
	return o&OptionStyleCheck != 0
}

// UncheckedAssert flags all unchecked asserts on errors.
func (o Options) UncheckedAssert() bool {
	return o&OptionUncheckedAssert != 0
}

// SetOption modifies the state of a specific option flag in the Options configuration
// based on the provided boolean value. The flag parameter should be a single option
// constant (e.g., [OptionCheckIs], [OptionStyleCheck]).
func (o *Options) SetOption(flag Options, v bool) {
	if v {
		*o |= flag
	} else {
		*o &^= flag
	}
}
