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

import "errors"

type zError struct{}

func (zError) Error() string {
	return "z"
}

// A type rename keeps the "Error" suffix: the count is inserted before it.
type zErr struct{} // want ` suggestion: "z2Error"$`

func (zErr) Error() string {
	return "z"
}

var ErrSame = errors.New("same")

var SameErr = errors.New("same too") // want ` suggestion: "ErrSame2"$`

type Error struct{}

func (Error) Error() string {
	return "error"
}

// The whole name is the "Error" suffix: the numbered variant keeps the visibility.
type Err struct{} // want ` suggestion: "E2Error"$`

func (Err) Error() string {
	return "e"
}

// An unexported type must not get an exported suggestion.
type err struct{} // want ` suggestion: "eError"$`

func (err) Error() string {
	return "e"
}
