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

package testdata

import (
	"fmt"
	"strconv"
)

// Errorgeneric[T] is a generic error.
type Errorgeneric[T interface { // want ` suggestion: "GenericError"$`
	error
	fmt.Stringer
}] struct{ v T }

func (e Errorgeneric[T]) Error() string { return "error: " + e.v.String() }

type errc string // want ` suggestion: "cError"$`

func (e errc) String() string { return string(e) }

func (e errc) Error() string { return e.String() }

type ErrD int // want ` suggestion: "DError"$`

func (e ErrD) String() string { return strconv.Itoa(int(e)) }

func (e ErrD) Error() string { return e.String() }

var (
	cErr   Errorgeneric[errc] = Errorgeneric[errc]{v: "a"} // want ` suggestion: "errC"$`
	Derror Errorgeneric[ErrD] = Errorgeneric[ErrD]{v: 1}   // want ` suggestion: "ErrD2"$`
)
