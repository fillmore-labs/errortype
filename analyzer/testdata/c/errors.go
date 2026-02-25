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

package c

import (
	"errors"
	. "errors"
	"os"

	errorsx "golang.org/x/exp/errors"
	"golang.org/x/xerrors"
)

type my1Error struct{}

func (my1Error) Error() string {
	return ""
}

func (my1Error) Is(_, err error) bool { // want ` \(et:sig\)$`
	return errors.Is(err, nil)
}

type myEmbeddedError struct {
	*my1Error
}

func wrap(err, target error) (error, error) {
	return err, target
}

func Errors() {
	_ = errors.Is(my1Error{}, &my1Error{}) // want "is false or undefined"

	_ = errors.Is(&my1Error{}, &my1Error{}) // want "is false or undefined"

	_ = errors.Is(&my1Error{}, &myEmbeddedError{}) // want "is false or undefined"

	_ = errors.Is(&struct{ my1Error }{}, &my1Error{}) // want "is false or undefined"

	_ = Is(&my1Error{}, &my1Error{}) // want "is false or undefined"

	_ = Is(wrap(&my1Error{}, &my1Error{}))

	var e my1Error
	_ = errors.Is(nil, &e)

	_ = errors.As(my1Error{}, &my1Error{})

	_ = errors.Join(&my1Error{}, &my1Error{})

	_ = errors.Unwrap(&my1Error{})

	_ = errorsx.Is(&my1Error{}, &my1Error{}) // want "is false or undefined"

	_ = xerrors.Is(func() error { // want "is false or undefined"
		return &myIsError{}
	}(), &my1Error{})

	(errors.Is(nil, &e)) // want ` \(et:unu\+\)$`
}

func Errors2() {
	errors := my1Error{}
	_ = errors.Is(&my1Error{}, &my1Error{})
}

type StructWithIsField struct {
	Is func(_, _ error) bool
}

func Errors3() {
	errors := StructWithIsField{Is: func(_, _ error) bool { return false }}

	_ = errors.Is(&my1Error{}, &my1Error{})
}

type myIsError struct{}

func (myIsError) Error() string {
	return "my error with is"
}

func myErrosAs[E error](err error) (e E, ok bool) {
	ok = errors.As(err, &e)
	return
}

func (myIsError) Is(err error) bool {
	_, ok := myErrosAs[*myIsError](err) // want "should only shallowly compare"
	return ok
}

type myUnwrapError struct{}

func (myUnwrapError) Error() string {
	return "my error with unwrap"
}

func (myUnwrapError) Unwrap() error {
	return os.ErrProcessDone
}

type myUnwrapArrayError struct{}

func (myUnwrapArrayError) Error() string {
	return "my error with unwrap"
}

func (myUnwrapArrayError) Unwrap() []error {
	return []error{os.ErrProcessDone}
}

func Errors4() {
	_ = errors.Is(&myIsError{}, &my1Error{})

	_ = errors.Is(&struct{ *myIsError }{}, &my1Error{})

	_ = errors.Is(&myUnwrapError{}, os.ErrProcessDone)

	_ = errors.Is(&myUnwrapArrayError{}, os.ErrProcessDone)

	_ = errors.Is(os.ErrProcessDone, &myUnwrapError{}) // want `type "?myUnwrapError"?`

	_ = errors.Is(UnnamedIsError{}, UnnamedIsError{})

	_ = errors.Is(UnderscoreIsError{}, UnderscoreIsError{})
}

func Interface() {
	var errors interface{ Is(err, target error) bool }

	_ = errors.Is(my1Error{}, &my1Error{})
}

func Index() {
	var Is [1]func(err, target error) bool

	_ = Is[0](my1Error{}, &my1Error{})
}

type comparableError struct{ _ [0]byte }

func (comparableError) Error() string {
	return ""
}

type nonComparableError struct{ _ [0][]byte }

func (nonComparableError) Error() string {
	return ""
}

type nonComparableWithIsError struct{ _ [0][]byte }

func (nonComparableWithIsError) Error() string {
	return ""
}

func (nonComparableWithIsError) Is(err error) bool {
	return errors.As(err, new(nonComparableWithIsError)) // want "only shallowly compare"
}

func NonComparable() {
	_ = errors.Is(comparableError{}, comparableError{})

	_ = errors.Is(nonComparableError{}, nonComparableError{}) // want "non-comparable .* is always false"

	_ = errors.Is(nonComparableWithIsError{}, nonComparableWithIsError{})
}

type UnnamedIsError struct{}

func (UnnamedIsError) Error() string { return "unnamed" }

func (UnnamedIsError) Is(error) bool {
	return false
}

type UnderscoreIsError struct{}

func (_ UnderscoreIsError) Error() string { return "underscore" }

func (_ UnderscoreIsError) Is(_ error) bool {
	return false
}
