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

package errortypes

// ErrorType represents a detected usage type of an error.
type ErrorType uint8

// Constants defining the possible usages of an error type.
const (
	PointerType ErrorType = 1 << iota // Pointer
	ValueType                         // Value

	SuppressType = PointerType | ValueType // Suppress

	Undecided ErrorType = 0 // Undecided

	ExpectedMask = PointerType | ValueType
)

func (e ErrorType) String() string {
	switch e & ExpectedMask {
	case Undecided:
		return "Undecided"

	case ValueType:
		return "Value"

	case PointerType:
		return "Pointer"

	case SuppressType:
		return "Suppress"
	}

	return "<invalid>"
}

// AFact makes *ErrorType satisfy the [analysis.Fact] interface.
// [analysis.Fact]s must be pointers to be exported as a fact.
func (*ErrorType) AFact() {}
