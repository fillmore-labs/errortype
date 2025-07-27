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

package a

import (
	"errors"
	. "errors"

	"test/a/b"

	pkgerrors "github.com/pkg/errors"
	errorsx "golang.org/x/exp/errors"
	"golang.org/x/xerrors"
)

var ErrOne, ErrTwo error

type myError1 struct{}

func (myError1) Error() string {
	return ""
}

func (myError1) As(_ error, _ any) bool {
	return false
}

func Errors() {
	_ = errors.As(&myError1{}, &b.AmbiguousError{}) // want " \\(et:emb\\)$" " \\(et:sty\\)$"

	_ = As(&myError1{}, &b.AmbiguousError{}) // want " \\(et:emb\\)$" " \\(et:sty\\)$"

	_ = xerrors.As(func() error {
		return &myErrorWithAs{} // want " \\(et:ret\\)$"
	}(), &b.AmbiguousError{}) // want " \\(et:emb\\)$" " \\(et:sty\\)$"

	_ = errorsx.As(&myError1{}, &b.AmbiguousError{}) // want " \\(et:emb\\)$" " \\(et:sty\\)$"

	_ = pkgerrors.As(&myError1{}, &b.AmbiguousError{}) // want " \\(et:emb\\)$" " \\(et:sty\\)$"
}

func Errors2() {
	errors := myError1{}
	_ = errors.As(&myError1{}, &b.AmbiguousError{})
}

type StructWithAsField struct {
	As func(_ error, _ any) bool
}

func Errors3() {
	errors := StructWithAsField{As: func(_ error, _ any) bool { return false }}

	_ = errors.As(&myError1{}, &b.AmbiguousError{})
}

type myErrorWithAs struct{}

func (myErrorWithAs) Error() string {
	return "my error with as"
}

var _ error = myErrorWithAs{}

func (m myErrorWithAs) As(target any) bool {
	var success bool

	if t, ok := target.(*myErrorWithAs); ok {
		*t = m
		success = true
	}

	if err, ok := target.(error); ok {
		_, _ = err.(*myErrorWithAs) // want " \\(et:ast\\)$"
	}

	return success
}

func Errors4() {
	var err error

	_, _ = err.(myErrorWithAs)

	_ = errors.As(myErrorWithAs{}, &b.AmbiguousError{}) // want " \\(et:emb\\)$" " \\(et:sty\\)$"

	_ = myErrorWithAs{}.As(&b.AmbiguousError{})
}
