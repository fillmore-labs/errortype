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

package style

import (
	"errors"
	"strconv"
)

type MyError struct{ v int }

func (e MyError) Error() string {
	return "my error " + strconv.Itoa(e.v)
}

var _ error = (*MyError)(nil)

func Style1(err error) bool {
	return errors.As(err, &MyError{}) // want " \\(et:err\\+\\)$" " \\(et:sty\\)$"
}

func Style2(err error) bool {
	if e := &(MyError{}); errors.As(err, e) { // want " \\(et:err\\+\\)$" " \\(et:sty\\)$"
		return true
	}

	return false
}

func Style3(err error) bool {
	if e := new(MyError); errors.As(err, e) { // want " \\(et:err\\+\\)$" " \\(et:sty\\)$"
		return true
	}

	return false
}

func Style4(err error) bool {
	e := MyError{}

	return errors.As(err, &e) // want " \\(et:err\\+\\)$"
}

func Style5(err error) bool {
	var e *MyError
	ep := &e

	return errors.As(err, *ep) // want " \\(et:err\\+\\)$" " \\(et:sty\\)$"
}

func StyleX(err error) bool {
	var e MyError

	return errors.As(err, &e) // want " \\(et:err\\+\\)$"
}
