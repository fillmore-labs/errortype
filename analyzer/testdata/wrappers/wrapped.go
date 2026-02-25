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
	"test/wrappers/iface"
	"test/wrappers/wrap"
)

func WrapAs(err error, target any) bool {
	return wrap.As(err, target)
}

func WrapIs(err, target error) bool {
	return wrap.Is(err, target)
}

func WrapAsType[E error](err error) (E, bool) {
	return wrap.AsType[E](err)
}

func TestSub(s iface.Suite) {
	var (
		err    error
		target MyError
	)

	_ = WrapIs(err, &MyError{}) // want ` \(et:cmp\)$`

	_ = WrapAs(err, target) // want ` \(et:arg\)$`

	_, _ = WrapAsType[*MyError](err) // want ` \(et:ast\)$`

	_ = s.Is(err, &MyError{}) // want ` \(et:cmp\)$`

	_ = s.As(err, target) // want ` \(et:arg\)$`
}
