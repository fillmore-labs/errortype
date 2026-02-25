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

// DocErr wraps a value with no other interesting properties.
type DocErr struct{} // want ` suggestion: "DocError"$`

func (DocErr) Error() string { return "doc" }

// Err8 and Err9 are errors.
type (
	Err8 struct{} // want ` suggestion: "E8Error"$`
	Err9 struct{} // want ` suggestion: "E9Error"$`
)

func (Err8) Error() string { return "err1" }

func (Err9) Error() string { return "err2" }
