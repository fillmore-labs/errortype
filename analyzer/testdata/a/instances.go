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

package a

import (
	"errors"

	"test/a/b"
)

type errBox[E error] struct{ e E }

type errAlias[E error] = struct{ e E }

type errPair[A any, E error] struct {
	a A
	e E
}

type compErr interface {
	comparable
	error
}

type customErr interface {
	error
	isCustom()
}

type customValError struct{ ValueError }

func (customValError) isCustom() {}

var _ error = customValError{}

func genericErrFunc[E error](_ E)      {}
func errPairFunc[A any, E error](_ A, _ E) {}
func twoErrsFunc[E1, E2 error](_ E1, _ E2) {}
func compErrFunc[E compErr](_ E)           {}
func customErrFunc[E customErr](_ E)       {}
func anyFunc[T any](_ T)                   {}
func compFunc[T comparable](_ T)           {}
func unionNoErrFunc[T *ValueError | int](_ T) {
}

func applyFunc[T any](fn func(T), v T) {
	fn(v)
}

func extraAsType[E1, E2 error](err error) (E1, bool) {
	return errors.AsType[E1](err)
}

func identityErr[E error](e E) E {
	return e
}

func identityAny[T any](v T) T {
	return v
}

type errMatcher[E error] struct{}

func (errMatcher[E]) AsType(err error) (E, bool) {
	return errors.AsType[E](err)
}

func TestInstances(err error) {
	// Generic struct instantiations
	var (
		_ errBox[*ValueError] // want ` \(et:ins\)$`
		_ errBox[my1Error]    // want ` \(et:ins\+\)$`
		_ errBox[ValueError]
		_ errBox[*my1Error]
		_ errBox[error]
		_ errBox[*b.ValueError]   // want ` \(et:ins\)$`
		_ errBox[b.AmbiguousError] // want ` \(et:emb\)$`
	)

	// Generic type alias instantiations
	var (
		_ errAlias[*ValueError] // want ` \(et:ins\)$`
		_ errAlias[my1Error]    // want ` \(et:ins\+\)$`
		_ errAlias[ValueError]
		_ errAlias[*my1Error]
	)

	// Generic function instantiations (explicit and inferred)
	genericErrFunc[*ValueError](&ValueError{}) // want ` \(et:ins\)$`
	genericErrFunc[my1Error](my1Error{})       // want ` \(et:ins\+\)$`
	genericErrFunc(ValueError{})
	genericErrFunc(&my1Error{})
	genericErrFunc(&ValueError{}) // want ` \(et:ins\)$`
	genericErrFunc(my1Error{})    // want ` \(et:ins\+\)$`

	// Multi-parameter generics and partial inference
	var _ errPair[int, *ValueError]               // want ` \(et:ins\)$`
	errPairFunc[int, *ValueError](1, &ValueError{}) // want ` \(et:ins\)$`
	errPairFunc[int](1, &ValueError{})             // want ` \(et:ins\)$`
	twoErrsFunc[*ValueError, my1Error](&ValueError{}, my1Error{}) // want ` \(et:ins\)$` ` \(et:ins\+\)$`

	// Non-basic error constraints
	compErrFunc(&ValueError{}) // want ` \(et:ins\)$`
	customErrFunc(&customValError{}) // want ` \(et:ins\)$`
	customErrFunc(customValError{})

	// Uncalled function instantiations, subsequent calls, and function-argument inference
	_ = genericErrFunc[*ValueError]          // want ` \(et:ins\)$`
	asVal := errors.AsType[*ValueError]      // want `should be queried as a value .* \(et:ast\)$`
	_, _ = asVal(err)                        // call is not an instantiation
	applyFunc(genericErrFunc, &ValueError{}) // want ` \(et:ins\)$`

	// Method call on an already-instantiated generic struct (no double diagnosis)
	var m errMatcher[*ValueError] // want ` \(et:ins\)$`
	_, _ = m.AsType(err)          // call is not an instantiation

	// AsType wrapper with an extra error-constrained type parameter (no double diagnosis)
	_, _ = extraAsType[*ValueError, my1Error](err) // want `should be queried as a value .* \(et:ast\)$` `should be instantiated as a pointer .* \(et:ins\+\)$`

	// Follow-up VarDecl suppression for error-constrained type parameter results
	valErr := identityErr[*ValueError](&ValueError{}) // want ` \(et:ins\)$`
	_ = valErr

	var errVal = identityErr(&ValueError{}) // want ` \(et:ins\)$`
	_ = errVal

	anyErr := identityAny[*ValueError](&ValueError{}) // want ` \(et:var\)$`
	_ = anyErr

	// Non-overlap with errorconv.CallArgs
	genericErrFunc[error](&ValueError{}) // want ` \(et:arg\)$`

	// Non-error constraints are ignored
	anyFunc(&ValueError{})
	compFunc(my1Error{})
	unionNoErrFunc(&ValueError{})
}
