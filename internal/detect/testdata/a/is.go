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

import "math/rand/v2"

type (
	ErrorWithIs1    struct{ error }
	ErrorWithIs2    struct{ error }
	ErrorWithoutIs1 struct{ error }
	ErrorWithoutIs2 struct{ error }
)

func (e *ErrorWithIs1) Is(err error) bool {
	return e.error == err
}

func (e ErrorWithIs2) Is(err error) bool {
	return e.error == err
}

func (e *ErrorWithoutIs1) Is(err1, err2 error) bool {
	return e.error == err1
}

func (e *ErrorWithoutIs2) Is(error) string {
	return e.error.Error()
}

func ErrorIs() error {
	switch rand.Int() {
	case 0:
		return ErrorWithIs1{} // want "POINTER"

	case 1:
		return ErrorWithIs2{} // want "VALUE"

	case 2:
		return ErrorWithoutIs1{} // want "VALUE"

	case 3:
		return ErrorWithoutIs2{} // want "VALUE"

	default:
		return nil
	}
}
