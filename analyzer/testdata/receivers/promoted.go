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

package receivers

import "fmt"

// A promoted Error is not reported...
type PromotedUnwrapError struct{ error }

func (e *PromotedUnwrapError) Unwrap() error { return e.error }

// ...but Unwrap makes it a pointer error, so using the value is.
var ErrPromotedUnwrap = PromotedUnwrapError{} // want ` \(et:var\+\)$`

// An assertion declares the value form, so the receiver of Unwrap is the mistake.
type AssertedValueError struct{ error }

var _ error = AssertedValueError{}

func (e *AssertedValueError) Unwrap() error { return e.error } // want `"Unwrap" should be implemented with a value receiver .* \(et:rcv\)$`

func GetAssertedValue(err error) error { return fmt.Errorf("get: %w", AssertedValueError{err}) }

// The same holds when Error is declared.
type AssertedDeclaredError struct{ err error }

var _ error = AssertedDeclaredError{}

func (e AssertedDeclaredError) Error() string { return e.err.Error() }

func (e *AssertedDeclaredError) Is(err error) bool { return e.err == err } // want `"Is" should be implemented with a value receiver .* \(et:rcv\)$`

// An assertion for the pointer agrees with Unwrap.
type AssertedPointerError struct{ error }

var _ error = (*AssertedPointerError)(nil)

func (e *AssertedPointerError) Unwrap() error { return e.error }

func GetAssertedPointer(err error) error {
	return fmt.Errorf("get: %w", AssertedPointerError{err}) // want ` \(et:wrp\+\)$`
}

// Contradictory assertions leave the decision to Unwrap.
type AssertedBothError struct{ error }

var (
	_ error = (*AssertedBothError)(nil)
	_ error = AssertedBothError{} // want ` \(et:var\+\)$`
)

func (e *AssertedBothError) Unwrap() error { return e.error }

var ErrAssertedBoth = AssertedBothError{} // want ` \(et:var\+\)$`

// There is no receiver to change when Unwrap is promoted, so the assertion is reported.
type unwrapper struct{ err error }

func (u *unwrapper) Unwrap() error { return u.err }

type AssertedPromotedUnwrapError struct {
	error
	unwrapper
}

var _ error = AssertedPromotedUnwrapError{} // want ` \(et:var\+\)$`
