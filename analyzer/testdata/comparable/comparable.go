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

package comparable

import "errors"

type myErrors []error

func (e myErrors) Error() string {
	err := errors.Join(e...)
	if err == nil {
		return "<nil>"
	}

	return err.Error()
}

func (e myErrors) Unwrap() []error {
	return e
}

type _ = myErrors

type MyAliasErrors = myErrors // want ` \(et:nce\+\)$`

// Deprecated: Use [MyErrors] instead.
type MyDeprecatedAliasErrors = myErrors

//nolint:errortype
type MyNolintAliasErrors = myErrors

type MyErrors struct{ errs []error } // want ` \(et:nce\+\)$`

func (e MyErrors) Error() string {
	err := errors.Join(e.errs...)
	if err == nil {
		return "<nil>"
	}

	return err.Error()
}

func (e MyErrors) Unwrap() []error {
	return e.errs
}

type myWithIsError string

func (e myWithIsError) Error() string { return string(e) }

func (e myWithIsError) Is(err error) bool { return err == e }

type MyUncomparableError struct {
	_ [0][]byte

	myWithIsError
}

var _ error = MyUncomparableError{}

type MyOtherUncomparableError struct { // want ` \(et:nce\+\)$`
	_ [0][]byte

	error
}

var _ error = MyOtherUncomparableError{}

var ErrMyUncomparableSentinel = MyOtherUncomparableError{} // want ` \(et:nce\)$`
