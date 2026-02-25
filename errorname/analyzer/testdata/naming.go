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
	"errors"
	"strconv"
)

type Errora struct{} // want ` suggestion: "AError"$`

func (e Errora) Error() string {
	return "error a"
}

func (e Errora) a() string {
	return e.a()
}

type Errorb struct{} // want ` suggestion: "BError"$`

func (e Errorb) Error() string {
	return e.b()
}

func (e Errorb) b() string {
	return "error b"
}

type MyErr struct{ _ int } // want ` suggestion: "MyError"$`

func (*MyErr) Error() string {
	return "my error"
}

var sentinelError = errors.New("sentinel error") // want ` suggestion: "errSentinel"$`

type NumErr int // want ` suggestion: "NumError"$`

func (e NumErr) Error() string {
	return "error " + strconv.Itoa(int(e))
}

const (
	NumErr0 NumErr = 0    // want ` suggestion: "ErrNumErr0"$`
	NumErr1 NumErr = iota // want ` suggestion: "ErrNumErr1"$`
	NumErr2               // want ` suggestion: "ErrNumErr2"$`
)

const (
	NumErr3 NumErr = 3         // want ` suggestion: "ErrNumErr3"$`
	NumErr4 NumErr = 4         // want ` suggestion: "ErrNumErr4"$`
	NumErr5        = NumErr(5) // want ` suggestion: "ErrNumErr5"$`
)
