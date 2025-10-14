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

type IntError = b.GenericError[int]

type StringError = b.GenericError[string]

type GenericError[T string | int] = b.GenericError[T]

var _, _, _ error = IntError{},
	&StringError{}, // want " \\(et:var\\)$"
	&GenericError[string]{} // want " \\(et:var\\)$"

func Generics() {
	var err error

	_, _ = err.(*IntError) // want " \\(et:ast\\)$"
	_, _ = err.(StringError)

	switch err.(type) {
	case *IntError: // want " \\(et:ast\\)$"
	case StringError:
	case nil:
	default:
	}

	switch err.(type) {
	case GenericError[int]:
	case *GenericError[int]: // want " \\(et:ast\\)$"
	case nil:
	default:
	}

	switch err.(type) {
	case *b.GenericError[int]: // want " \\(et:ast\\)$"
	case b.GenericError[string]:
	case nil:
	default:
	}

	_ = func() error {
		return (*IntError)(nil) // want " \\(et:ret\\)$"
	}

	_ = func() error {
		return StringError{}
	}

	_ = func() error {
		return GenericError[int]{}
	}

	_ = func() error {
		return &GenericError[int]{} // want " \\(et:ret\\)$"
	}

	_ = func() error {
		return (*IntError)(&b.GenericError[int]{}) // want " \\(et:ret\\)$"
	}

	_ = func() error {
		return StringError(b.GenericError[string]{})
	}

	_ = func() error {
		return &b.GenericError[int]{} // want " \\(et:ret\\)$"
	}

	_ = func() error {
		return b.GenericError[string]{}
	}
}
