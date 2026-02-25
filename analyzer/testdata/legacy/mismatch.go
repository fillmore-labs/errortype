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

package legacy

import "errors"

type ValueError struct{ msg string }

func (e ValueError) Error() string { return e.msg }

var _ error = ValueError{}

type PointerError struct{ msg string }

func (e PointerError) Error() string { return e.msg }

var _ error = (*PointerError)(nil)

type pointerAlias = *ValueError

var lastValue *ValueError

// The fix queries the detected form, since the asserted form never matches.

func IsValueError(err error) bool { // want ` \(et:lgc\)$`
	_, ok := err.(*ValueError) // want ` \(et:ast\)$`
	return ok
}

func IsPointerError(err error) bool { // want ` \(et:lgc\)$`
	_, ok := err.(PointerError) // want ` \(et:ast\+\)$`
	return ok
}

// The fix keeps the asserted form: the value is stored in a variable of that type.

func IsValueErrorStored(err error) (ok bool) { // want ` \(et:lgc\)$`
	lastValue, ok = err.(*ValueError) // want ` \(et:ast\)$`
	return ok
}

// The fix keeps the asserted form: the alias can't be rewritten syntactically.

func IsValueErrorAlias(err error) bool { // want ` \(et:lgc\)$`
	_, ok := err.(pointerAlias) // want ` \(et:ast\)$`
	return ok
}

var _ = errors.New
