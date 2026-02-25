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

import "fmt"

type (
	errmy1 int // want ` suggestion: "my1Error"$`
	My2Err int // want ` suggestion: "My2Error"$`
)

func (e errmy1) Error() string {
	return fmt.Sprintf("my error %d", e)
}

func (e My2Err) Error() string {
	return fmt.Sprintf("my error %d", e)
}

var (
	my1 errmy1 = 1 // want ` suggestion: "errMy1"$`
	My2 My2Err = 2 // want ` suggestion: "ErrMy2"$`
)

const (
	My3 errmy1 = iota + 3 // want ` suggestion: "ErrMy3"$`
	My4                   // want ` suggestion: "ErrMy4"$`
	my5 My2Err = iota + 3 // want ` suggestion: "errMy5"$`
)

// myErr is unexported, but used in a generated file.
var myErr errmy1 // want ` suggestion: "errMy"$`

// errGen is unexported and used in a generated file,
// so it keeps its name behind an alias.
type errGen struct{ msg string } // want ` suggestion: "genError"$`

func (e *errGen) Error() string { return e.msg }
