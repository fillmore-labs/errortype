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

type myError1 struct{}

func (myError1) Error() string {
	return ""
}

type myErrorWithIs struct{ err error }

func (e myErrorWithIs) Error() string {
	return e.err.Error()
}

func (e myErrorWithIs) Is(err error) bool {
	return errors.Is(e.err, (err)) // want "only shallowly compare"
}

type myErrorWithUnwrap struct{ err error }

func (myErrorWithUnwrap) Error() string {
	return "my error with unwrap"
}

func (e myErrorWithUnwrap) Unwrap() error {
	return e.err
}

type myErrorWithUnwrapArray struct{ errs []error }

func (myErrorWithUnwrapArray) Error() string {
	return "my error with unwrap"
}

func (e myErrorWithUnwrapArray) Unwrap() []error {
	return e.errs
}

func Errors4() {
	_ = errors.Is(&myErrorWithIs{}, &myError1{}) // want "is always false"

	_ = errors.Is(&struct{ *myErrorWithIs }{}, &myError1{}) // want "is always false"

	_ = errors.Is(&myErrorWithUnwrap{}, os.ErrProcessDone) // want "is always false"

	_ = errors.Is(&myErrorWithUnwrapArray{}, os.ErrProcessDone) // want "is always false"
}

type nonError struct{ err error }

func (e nonError) Error() []byte {
	return []byte(e.err.Error())
}

func (e nonError) Is(err error) bool {
	return errors.Is(e.err, (err))
}
