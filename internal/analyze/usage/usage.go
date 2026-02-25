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

package usage

import "fillmore-labs.com/errortype/detect/result"

//go:generate go tool bitmask -type Usage

// Usage represents the expected and observed usage of an error type.
type Usage uint8

// Constants defining the possible usages of an error type.
const (
	// PointerExpected indicates the error type should be used as a pointer ("&MyError{}").
	PointerExpected Usage = 1 << iota // PointerExpected

	// ValueExpected indicates the error type should be used as a value ("MyError{}").
	ValueExpected // ValueExpected

	// PointerObserved is set when a pointer usage was observed.
	PointerObserved // PointerObserved

	// ValueObserved is set when a value usage was observed.
	ValueObserved // ValueObserved

	// SuppressExpected indicates that analysis for this error type should be suppressed.
	SuppressExpected = PointerExpected | ValueExpected

	// ExpectedMask selects the ...Expected usages.
	ExpectedMask = PointerExpected | ValueExpected

	// ObservedMask selects the ...Observed usages.
	ObservedMask = PointerObserved | ValueObserved
)

// DeterminedType analyzes the usage pattern of the Usage value and determines
// if there is an observed usage type (consistent pointer, consistent value, or mixed)
// that differs from the expected analysis type.
func (u Usage) DeterminedType() result.ErrorType {
	// Do we have a consistent use that is different from the detected type?
	switch u & ObservedMask {
	case PointerObserved:
		if u&ExpectedMask != PointerExpected {
			return result.Pointer
		}

	case ValueObserved:
		if u&ExpectedMask != ValueExpected {
			return result.Value
		}

	case PointerObserved | ValueObserved:
		if u&ExpectedMask != SuppressExpected {
			return result.Suppress
		}
	}

	return result.Undecided
}
