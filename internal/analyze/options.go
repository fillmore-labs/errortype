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

import "strconv"

//go:generate go tool bitmask -type Options -boolflag

// Options represent configuration flags to control the behavior of style and correctness checks for errors.
type Options uint16

const (
	// OptionCheckIs controls whether to check for `Is(error) bool` methods.
	OptionCheckIs Options = 1 << iota // check-is
	// OptionCheckUnused flags unchecked results of `errors.Is` calls.
	OptionCheckUnused // check-unused
	// OptionDeepIsCheck flags all unwrap methods in `Is` method checks, not only those on target.
	OptionDeepIsCheck // deep-is-check
	// OptionStyleCheck controls the target style check in `errors.As` calls.
	OptionStyleCheck // style-check
	// OptionUncheckedAssert flags all unchecked asserts on errors.
	OptionUncheckedAssert // unchecked-assert
	// OptionPrefixFilter restricts variable declaration analysis to variables with an 'err' or 'Err' prefix.
	OptionPrefixFilter // prefix-filter

	// OptionNotComparable has extended checks for not comparable types.
	OptionNotComparable // non-comparable

	// OptionNaming checks sentinel and error type naming.
	OptionNaming // naming

	// OptionLegacy checks for pre-Go 1.13 legacy error assertion queries.
	OptionLegacy // legacy

	// OptionGenerated diagnoses in generated files, too.
	OptionGenerated // generated

	// DefaultOptions is the default configuration for error analysis.
	DefaultOptions = OptionCheckIs | OptionCheckUnused | OptionPrefixFilter | OptionGenerated

	// RecommendedOptions are additional recommended option.
	RecommendedOptions = OptionStyleCheck | OptionNotComparable | OptionNaming | OptionLegacy
)

// Recommended is a function to set recommended options with [flag.BoolFunc].
func (o *Options) Recommended(s string) error {
	value, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}

	if value { // redommended=false is a no-op.
		o.Set(RecommendedOptions, true)
	}

	return nil
}
