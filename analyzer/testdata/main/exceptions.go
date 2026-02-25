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

package main

import "errors"

type myError struct{ Msg string }

func (m *myError) Error() string { return m.Msg }

func (m *myError) f() (*myError, bool) { return m, true }

type myEmbeddedError struct{ *myError }

func Exception1() {
	var err error = myEmbeddedError{&myError{Msg: "embedded"}} // want ` \(et:emb\)$`

	var _ error = &myEmbeddedError{} // want ` \(et:emb\+\)$`

	var emb myEmbeddedError

	_ = errors.As(err, &emb) // want ` \(et:emb\)$`

	var embp *myEmbeddedError

	_ = errors.As(err, &embp) // want ` \(et:emb\+\)$`

	_, _ = err.(*myError).f() // want ` \(et:uca\)$`
}

type myInterfaceError interface{ error }

func Exception2() {
	emb := myEmbeddedError{&myError{Msg: "embedded"}}

	_ = &myEmbeddedError{}

	var err error = emb // want ` \(et:emb\)$`

	var myi myInterfaceError

	_ = err.(myInterfaceError)

	_ = err.(any)

	_ = err.(interface{ Unwrap() error }) // want ` \(et:uca\+\)$`

	_ = errors.As(err, &myi)
}
