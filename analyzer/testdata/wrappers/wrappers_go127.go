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

//go:build go1.27

package wrappers

import (
	"errors"

	"test/wrappers/wrap"
)

type Suite[S any] struct{}

func (s Suite[S]) Is[E error](err, target error) bool {
	return errors.Is(err, target)
}

func (s Suite[A]) AsType[T any, E error](err error) (E, bool) {
	return errors.AsType[E](err)
}

func Test127(s Suite[string], err error) {
	_ = s.Is[*MyError](err, &MyError{}) // want ` \(et:cmp\)$`

	_, _ = s.AsType[string, *MyError](err) // want ` \(et:ast\)$`

	_ = (Suite[string]).Is[*MyError](s, err, &MyError{}) // want ` \(et:cmp\)$`

	_, _ = (Suite[string]).AsType[string, *MyError](s, err) // want ` \(et:ast\)$`
}

func Test127Generic[T any, E error](s Suite[T], err error) {
	_ = s.Is[E](err, &MyError{}) // want ` \(et:cmp\)$`

	_, _ = s.AsType[T, *MyError](err) // want ` \(et:ast\)$`

	_ = (Suite[T]).Is[E](s, err, &MyError{}) // want ` \(et:cmp\)$`

	_, _ = (Suite[T]).AsType[T, *MyError](s, err) // want ` \(et:ast\)$`
}

func Test127Wrap(s wrap.Suite[string], err error) {
	_ = s.Is[*MyError](err, &MyError{}) // want ` \(et:cmp\)$`

	_, _ = s.AsType[string, *MyError](err) // want ` \(et:ast\)$`

	_ = (wrap.Suite[string]).Is[*MyError](s, err, &MyError{}) // want ` \(et:cmp\)$`

	_, _ = (wrap.Suite[string]).AsType[string, *MyError](s, err) // want ` \(et:ast\)$`
}

func Test127WrapGeneric[T any, E error](s wrap.Suite[T], err error) {
	_ = s.Is[E](err, &MyError{}) // want ` \(et:cmp\)$`

	_, _ = s.AsType[T, *MyError](err) // want ` \(et:ast\)$`

	_ = (wrap.Suite[T]).Is[E](s, err, &MyError{}) // want ` \(et:cmp\)$`

	_, _ = (wrap.Suite[T]).AsType[T, *MyError](s, err) // want ` \(et:ast\)$`
}
