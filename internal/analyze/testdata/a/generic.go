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

import "github.com/go-errors/errors"

type GenericError[T any] struct{ _ T }

func (GenericError[T]) Error() string { return "" }

type IntError = GenericError[int]

type StringError = GenericError[string]

var _, _ error = IntError{}, &StringError{}

func Generics() {
	var err error

	_, _ = err.(*IntError)   // want " \\(et:ast\\)$"
	_, _ = err.(StringError) // want " \\(et:ast\\+\\)$"

	switch err.(type) {
	case *IntError: // want " \\(et:ast\\)$"
	case StringError: // want " \\(et:ast\\+\\)$"
	case nil:
	default:
	}

	var itarget *IntError
	_ = errors.As(err, &itarget) // want " \\(et:err\\)$"

	var starget StringError
	_ = errors.As(err, &starget) // want " \\(et:err\\+\\)$"

	_ = func() error {
		return (*IntError)(nil) // want " \\(et:ret\\)$"
	}

	_ = func() error {
		return StringError{} // want " \\(et:ret\\+\\)$"
	}

	_ = func() error {
		return IntError(GenericError[int]{})
	}

	_ = func() error {
		return (*StringError)(&GenericError[string]{})
	}
}
