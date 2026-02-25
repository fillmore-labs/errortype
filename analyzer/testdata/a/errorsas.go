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
	"fmt"
)

type AsValueError struct{}

func (AsValueError) Error() string { return "" }

var _ error = AsValueError{}

type AsPointerError struct{ _ int }

func (AsPointerError) Error() string { return "" }

func NewErrorsAsPointer() interface{ error } { return &AsPointerError{} }

func ErrorsAs(err error) {
	var (
		evv AsValueError
		evp *AsValueError
		epv AsPointerError
		epp *AsPointerError

		eany interface{}
	)

	_ = errors.As(err, &evv)
	_ = errors.As(err, &evp) // want ` \(et:err\)$`
	_ = errors.As(err, &epv) // want ` \(et:err\+\)$`
	_ = errors.As(err, &epp)

	_ = errors.As(err, &AsValueError{})   // want ` \(et:sty\)$`
	_ = errors.As(err, &AsPointerError{}) // want ` \(et:err\+\)$` ` \(et:sty\)$`

	_ = errors.As(err, eany)
	_ = errors.As(err, &eany)

	_ = errors.As(func() (error, any) { return nil, nil }())

	_ = errors.As(err, evv) // want ` \(et:arg\)$`
	_ = errors.As(err, evp) // want ` \(et:sty\)$`
	_ = errors.As(err, epv) // want ` \(et:arg\)$`
	_ = errors.As(err, epp) // want ` \(et:err\+\)$` ` \(et:sty\)$`

	// Those are readable
	_ = errors.As(err, new(AsValueError))
	_ = errors.As(err, new(*AsPointerError))

	// Those are not
	_ = errors.As(err, &AsValueError{})   // want ` \(et:sty\)$`
	_ = errors.As(err, &AsPointerError{}) // want ` \(et:err\+\)$` ` \(et:sty\)$`
}

type TestError struct{ msg string }

func (t TestError) Error() string { return t.msg }
func (*TestError) F()             {}

func TestErrorAs() {
	var err TestError
	// While valid, we flag this construct
	var target interface{ F() } = &err
	if errors.As(TestError{msg: "hello"}, target) { // want ` \(et:arg\+\)$`
		fmt.Println(err.Error())
	}
}
