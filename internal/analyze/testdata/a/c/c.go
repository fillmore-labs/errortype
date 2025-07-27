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

type (
	C1 struct{}

	C1a = C1

	C2 struct{}

	C2a = C2

	C2p = *C2a

	C3 struct{}

	C4 struct{}

	C5 struct{}

	C6 struct{}

	C7 struct{}

	MyString = string
)

func (C1a) Error() string { return "" }

func (C2p) Error() MyString { return MyString("") }

func (C3) error() string { return "" }

func (C4) Error() bool { return false }

func (C5) Error() {}

func (C6) Error() (string, string) { return "", "" }

func (C7) Error(_ error) string { return "" }

func Error() string { return "" }

var (
	_, _ error = C1a{}, new(C2a)

	_ error = new(C2a)

	_ error = &struct{ C2a }{}

	_ error = struct{ error }{error: nil}

	_ error = nil

	_ error = error(nil)

	_ error

	_ = C3{}

	_ = C4{}

	_ = C5{}
)
