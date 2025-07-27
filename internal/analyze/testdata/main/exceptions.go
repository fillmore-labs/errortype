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

type myErr struct{ Msg string }

func (m *myErr) Error() string { return m.Msg }

type myErrorEmbedded struct{ *myErr }

func Exception1() {
	var err error = myErrorEmbedded{&myErr{Msg: "embedded"}}

	var _ error = &myErrorEmbedded{}

	var emb myErrorEmbedded

	_ = errors.As(err, &emb) // want " \\(et:emb\\)$"

	var embp *myErrorEmbedded

	_ = errors.As(err, &embp) // want " \\(et:emb\\+\\)$"
}

type myInterface interface{ error }

func Exception2() {
	emb := myErrorEmbedded{&myErr{Msg: "embedded"}}

	_ = &myErrorEmbedded{}

	var err error = emb

	var myi myInterface

	_ = err.(myInterface)

	_ = errors.As(err, &myi)
}
