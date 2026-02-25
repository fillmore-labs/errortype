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

package main

import "errors"

type valueTargetError struct{ _ int }

func (valueTargetError) Error() string { return "" }

var _ error = valueTargetError{}

// A pointer alias of a value error is still a pointer.
type valuePtrAliasError = *valueTargetError

func pointerAliasAssert(err error) bool {
	_, ok := err.(valuePtrAliasError) // want ` \(et:ast\)$`
	return ok
}

func pointerAliasAsType(err error) bool {
	_, ok := errors.AsType[valuePtrAliasError](err) // want ` \(et:ast\)$`
	return ok
}

func pointerAliasAs(err error) bool {
	var target valuePtrAliasError
	return errors.As(err, &target) // want ` \(et:err\)$`
}

func pointerAliasSwitch(err error) bool {
	switch err.(type) {
	case valuePtrAliasError: // want ` \(et:ast\)$`
		return true
	}
	return false
}

func pointerAliasReturn() error {
	return valuePtrAliasError(&valueTargetError{}) // want ` \(et:ret\)$`
}
