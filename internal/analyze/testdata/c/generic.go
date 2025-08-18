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

package c

import (
	"errors"
	"fmt"
)

type genericError[T fmt.Stringer] struct {
	e T
}

func (g genericError[T]) Error() string {
	return "error: " + g.e.String()
}

type empty struct{}

func (empty) String() string { return "empty" }

func Generic[T fmt.Stringer]() {
	_ = errors.Is(&genericError[T]{}, &genericError[T]{}) // want "is always false"

	_ = errors.Is(&genericError[empty]{}, &genericError[empty]{}) // want "is false or undefined"
}

func Generic2[T, U fmt.Stringer]() {
	Generic[T]()

	Generic2[T, U]()
}
