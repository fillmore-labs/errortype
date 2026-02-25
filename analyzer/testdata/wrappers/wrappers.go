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

package wrappers

import (
	"errors"
	"fmt"
)

func AsType[E error](err error) (E, bool) {
	return errors.AsType[E](err)
}

func TestAsType() {
	AsType[*MyError](nil) // want ` \(et:ast\)$` ` \(et:unu\)$`
}

func Errorf(format string, a ...any) error {
	return fmt.Errorf(format, a...)
}

func CustomErrorf(format string, a ...any) MyError {
	_ = fmt.Errorf(format, a...)

	return MyError{}
}

type sender struct{ ch chan<- error }

func (s sender) SendErrorf(format string, a ...any) bool {
	return SendErrorf(s.ch, format, a...)
}

func SendErrorf(ch chan<- error, format string, a ...any) bool {
	ch <- fmt.Errorf(format, a...)

	return true
}

func TestErrorf() {
	Errorf("nil: %w", nil)       // want ` \(et:unu\)$`
	CustomErrorf("nil: %w", nil) // want ` \(et:unu\)$`

	ch := make(chan error)
	s := sender{ch: ch}
	s.SendErrorf("nil: %w", &MyError{}) // want ` \(et:wrp\)$`

	SendErrorf(ch, "nil: %w", &MyError{}) // want ` \(et:wrp\)$`
}

func TestUnwrap() {
	errors.Unwrap(nil) // want ` \(et:unu\)$`
}
