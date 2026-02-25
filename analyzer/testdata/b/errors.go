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

package b

import (
	"errors"
	"os"
)

type my1Error struct{}

func (my1Error) Error() string {
	return ""
}

type myIsError struct{ err error }

func (e myIsError) Error() string {
	return e.err.Error()
}

func (e myIsError) Is(err error) bool {
	return errors.Is(e.err, (err)) // want "only shallowly compare"
}

type myUnwrapError struct{ err error }

func (myUnwrapError) Error() string {
	return "my error with unwrap"
}

func (e myUnwrapError) Unwrap() error {
	return e.err
}

type myUnwrapArrayError struct{ errs []error }

func (myUnwrapArrayError) Error() string {
	return "my error with unwrap"
}

func (e myUnwrapArrayError) Unwrap() []error {
	return e.errs
}

func Errors4(err error) {
	_ = errors.Is(&myIsError{}, &my1Error{}) // want "is always false"

	_ = errors.Is(&struct{ *myIsError }{}, &my1Error{}) // want "is always false"

	_ = errors.Is(&myUnwrapError{}, os.ErrProcessDone) // want "is always false"

	_ = errors.Is(&myUnwrapArrayError{}, os.ErrProcessDone) // want "is always false"

	(errors.Is(err, nil)) // want ` \(et:unu\+\)$`

	(errors.As(err, nil)) // want ` \(et:arg\)$` ` \(et:unu\)$`
}

type nonError struct{ err error }

func (e nonError) Error() []byte {
	return []byte(e.err.Error())
}

func (e nonError) Is(err error) bool {
	return errors.Is(e.err, (err))
}

func Is() {}
