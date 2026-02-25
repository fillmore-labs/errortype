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

import "errors"

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target any) bool {
	return errors.As(err, target)
}

func AsType[E error](err error) (E, bool) {
	var target E
	ok := errors.As(err, &target)

	return target, ok
}

func HasType[E error](err error) bool {
	return errors.As(err, new(E))
}

type MyError struct{}

func (MyError) Error() string { return "my error" }

var _ error = MyError{}

func Test() {
	var err error
	var target MyError

	_ = Is(err, &MyError{}) // want ` \(et:cmp\)$`

	_ = As(err, target) // want ` \(et:arg\)$`

	_, _ = AsType[*MyError](err) // want ` \(et:ast\)$`

	_ = HasType[*MyError](err) // want ` \(et:ast\)$`
}
