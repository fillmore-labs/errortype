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

import "errors"

type ErrorsAsValue struct{}

func (ErrorsAsValue) Error() string { return "" }

var _ error = ErrorsAsValue{}

type ErrorsAsPointer struct{ _ int }

func (ErrorsAsPointer) Error() string { return "" }

func NewErrorsAsPointer() interface{ error } { return &ErrorsAsPointer{} }

func ErrorsAs(err error) {
	var (
		evv ErrorsAsValue
		evp *ErrorsAsValue
		epv ErrorsAsPointer
		epp *ErrorsAsPointer

		eany interface{}
	)

	_ = errors.As(err, &evv)
	_ = errors.As(err, &evp) // want " \\(et:err\\)$"
	_ = errors.As(err, &epv) // want " \\(et:err\\+\\)$"
	_ = errors.As(err, &epp)

	_ = errors.As(err, &ErrorsAsValue{})   // want " \\(et:sty\\)$"
	_ = errors.As(err, &ErrorsAsPointer{}) // want " \\(et:err\\+\\)$" " \\(et:sty\\)$"

	_ = errors.As(err, eany)
}
