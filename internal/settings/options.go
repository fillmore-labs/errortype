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

package settings

import "strconv"

// Options represent configuration flags to control the behavior of style and correctness checks for errors.
//
//go:generate go tool bitmask -type Options -boolflag
type Options uint16

const (
	// OptionStrictIsTarget reports errors.Is calls with a new address even when an `Is(error) bool`
	// method could match it or the new value unwraps.
	OptionStrictIsTarget Options = 1 << iota // strict-is-target
	// OptionCheckUnused flags unused results of `errors.Is/As/AsType`, `fmt.Errorf` and their wrappers.
	OptionCheckUnused // check-unused
	// OptionStrictIsMethod flags all unwrapping calls in `Is` methods, not only those on target.
	OptionStrictIsMethod // strict-is-method
	// OptionStyleCheck controls the target style check in `errors.As` calls.
	OptionStyleCheck // style-check
	// OptionUncheckedAssert flags all unchecked asserts on errors.
	OptionUncheckedAssert // unchecked-assert
	// OptionPrefixFilter restricts variable declaration analysis to variables named like errors
	// (an "err" prefix or "Err" suffix, or an "Err" prefix when exported).
	OptionPrefixFilter // prefix-filter

	// OptionNonComparable has extended checks for non-comparable types.
	OptionNonComparable // non-comparable

	// OptionNaming checks sentinel and error type naming.
	OptionNaming // naming

	// OptionLegacy checks for pre-Go 1.13 legacy error assertion queries.
	OptionLegacy // legacy

	// OptionGenerated diagnoses in generated files, too.
	OptionGenerated // generated

	// OptionReceivers checks that a pointer error has a pointer receiver on Error, and a value error
	// value receivers on its exported methods.
	OptionReceivers // receivers

	// DefaultOptions is the default configuration for error analysis.
	DefaultOptions = OptionCheckUnused | OptionPrefixFilter | OptionGenerated

	// RecommendedOptions are additional recommended options.
	RecommendedOptions = OptionStyleCheck | OptionNonComparable | OptionNaming | OptionLegacy | OptionReceivers
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
