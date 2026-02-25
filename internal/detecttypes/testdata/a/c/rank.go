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

package c

import "errors"

// Conversions to error interfaces outrank construction evidence.
type (
	RankLiteralError struct{ error } // want RankLiteralError:"value"
	RankCastError    int             // want RankCastError:"pointer"
)

func (RankCastError) Error() string { return "" }

func _(errs chan<- error) {
	e := &RankLiteralError{} // pointer literal
	errs <- *e               // value conversion
}

func _() {
	var err error

	_ = RankCastError(1)     // value cast
	err = new(RankCastError) // pointer conversion

	_ = err
}

// Compile-time assertions outrank sentinel errors.
type (
	RankDeclPointerError struct{}        // want RankDeclPointerError:"pointer"
	RankDeclValueError   struct{}        // want RankDeclValueError:"value"
	RankDeclSplitError   struct{ _ int } // want RankDeclSplitError:"value"
)

func (RankDeclPointerError) Error() string { return "" }

func (RankDeclValueError) Error() string { return "" }

func (RankDeclSplitError) Error() string { return "" }

var (
	_                  error = (*RankDeclPointerError)(nil) // pointer assertion
	ErrRankDeclPointer error = RankDeclPointerError{}       // value sentinel
	_                  error = RankDeclValueError{}         // value assertion
	ErrRankDeclValue         = &RankDeclValueError{}        // pointer sentinel
	_, _               error = RankDeclSplitError{}, (*RankDeclSplitError)(nil)
	ErrRankDeclSplit         = RankDeclSplitError{} // contradictory assertions, decided by sentinel
)

// Generic instantiation ranking and no double-counting of AsType wrappers.
type (
	RankAsTypeTargetError    struct{ error } // want RankAsTypeTargetError:"pointer"
	RankAsTypeNoDoubleError  struct{ error } // want RankAsTypeNoDoubleError:"value"
	RankAssignInstanceError  struct{ error } // want RankAssignInstanceError:"pointer"
	RankInstanceLitError     struct{ error } // want RankInstanceLitError:"pointer"
	RankUncalledAsTypeError  struct{ error } // want RankUncalledAsTypeError:"pointer"
	RankAliasInstanceError   struct{ error } // want RankAliasInstanceError:"pointer"
	RankInferredAsTypeError  struct{ error } // want RankInferredAsTypeError:"value"
	RankMatcherInstanceError struct{ error } // want RankMatcherInstanceError:"pointer"
)

type errAlias[E error] = struct{ err E }

func wrapAsTypeOut[E error](err error, target *E) (ok bool) {
	*target, ok = errors.AsType[E](err)

	return ok
}

type errMatcher[E error] struct{}

func (errMatcher[E]) AsType(err error) (E, bool) {
	return errors.AsType[E](err)
}

func _(err error, p *RankAssignInstanceError) {
	// errors.AsType target outranks general generic instantiation.
	_, _ = errors.AsType[*RankAsTypeTargetError](err) // PointerTarget
	var _ errBox[RankAsTypeTargetError]               // ValueInstance

	// AsType is not double-counted as an instance: contradictory Targets fall through to Literal.
	var valTarget RankAsTypeNoDoubleError
	_ = errors.As(err, &valTarget)                      // ValueTarget
	_, _ = errors.AsType[*RankAsTypeNoDoubleError](err) // PointerTarget only
	_ = RankAsTypeNoDoubleError{}                       // ValueLiteral

	// Conversions to error outrank generic instantiations.
	err = p                               // PointerAssign
	var _ errBox[RankAssignInstanceError] // ValueInstance

	// Generic instantiations outrank literals.
	var _ errBox[*RankInstanceLitError] // PointerInstance
	_ = RankInstanceLitError{}          // ValueLiteral

	// Uncalled AsType instantiation counts as Target (even when invoked later) and outranks general Instance.
	fn := errors.AsType[*RankUncalledAsTypeError] // PointerTarget
	_, _ = fn(err)                                // call is not an instantiation
	var _ errBox[RankUncalledAsTypeError]         // ValueInstance

	// Inferred WrapperAsType instantiation counts as Target and outranks general Instance.
	var inferredTarget RankInferredAsTypeError
	_ = wrapAsTypeOut(err, &inferredTarget) // ValueTarget (inferred)
	var _ errBox[*RankInferredAsTypeError]  // PointerInstance

	// Method call on an instantiated generic struct does not double-count.
	var m errMatcher[*RankMatcherInstanceError] // PointerInstance
	_, _ = m.AsType(err)                        // call is not an instantiation
	_ = RankMatcherInstanceError{}              // ValueLiteral

	// Generic alias instantiation outranks literals.
	var _ errAlias[*RankAliasInstanceError] // PointerInstance
	_ = RankAliasInstanceError{}            // ValueLiteral

	_ = err
}
