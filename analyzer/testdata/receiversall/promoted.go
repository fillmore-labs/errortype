// Copyright 2026 Oliver Eikemeier. All Rights Reserved.
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

package receiversall

import "fmt"

// Nobody chose the receiver of a promoted Error, so it is not reported as long as the type is used consistently...
type PromotedError struct{ error }

func GetPromoted(err error) error { return fmt.Errorf("get: %w", &PromotedError{err}) }

type PromotedStructError struct{ ValueError }

var ErrPromotedStruct = &PromotedStructError{}

type PromotedIndirectError struct{ *PointerError }

var ErrPromotedIndirect = &PromotedIndirectError{}

// ...also with a pointer Unwrap, Is or As, which the value lacks.
type PromotedUnwrapError struct{ error }

func (e *PromotedUnwrapError) Unwrap() error { return e.error }

func GetPromotedUnwrap(err error) error { return fmt.Errorf("get: %w", &PromotedUnwrapError{err}) }

type PromotedIsError struct{ error }

func (*PromotedIsError) Is(error) bool { return false }

type PromotedAsError struct{ error }

func (*PromotedAsError) As(any) bool { return false }

// Using the value is reported.
func GetPromotedUnwrapValue(err error) error {
	return fmt.Errorf("get: %w", PromotedUnwrapError{err}) // want ` \(et:wrp\+\)$`
}

type AssertedError struct{ error }

var _ error = (*AssertedError)(nil)

var ErrAsserted = AssertedError{} // want ` \(et:var\+\)$`

// Value receivers are fine.
type PromotedValueError struct{ error }

func (e PromotedValueError) Unwrap() error { return e.error }

var ErrPromotedValue error = PromotedValueError{}
