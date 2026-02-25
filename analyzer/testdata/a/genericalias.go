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

import "test/a/b"

type AIntError = b.GenericError[int]

type AStringError = b.GenericError[string]

type AGenericError[T string | int] = b.GenericError[T]

var _, _, _ error = AIntError{},
	&AStringError{}, // want ` \(et:var\)$`
	&AGenericError[string]{} // want ` \(et:var\)$`

func AGenerics() {
	var err error

	_, _ = err.(*AIntError) // want ` \(et:ast\)$`
	_, _ = err.(AStringError)

	switch err.(type) {
	case *AIntError: // want ` \(et:ast\)$`
	case AStringError:
	case nil:
	default:
	}

	switch err.(type) {
	case AGenericError[int]:
	case *AGenericError[int]: // want ` \(et:ast\)$`
	case nil:
	default:
	}

	switch err.(type) {
	case *b.GenericError[int]: // want ` \(et:ast\)$`
	case b.GenericError[string]:
	case nil:
	default:
	}

	_ = func() error {
		return (*AIntError)(nil) // want ` \(et:ret\)$`
	}

	_ = func() error {
		return AStringError{}
	}

	_ = func() error {
		return AGenericError[int]{}
	}

	_ = func() error {
		return &AGenericError[int]{} // want ` \(et:ret\)$`
	}

	_ = func() error {
		return (*AIntError)(&b.GenericError[int]{}) // want ` \(et:ret\)$`
	}

	_ = func() error {
		return AStringError(b.GenericError[string]{})
	}

	_ = func() error {
		return &b.GenericError[int]{} // want ` \(et:ret\)$`
	}

	_ = func() error {
		return b.GenericError[string]{}
	}
}
