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

import (
	"math/rand/v2"

	"test/a/c"
)

type Interface interface{ error }

func NewInterface() Interface { return nil }

func NewStruct() error { return &struct{ error }{} }

func NewValueDefault() *c.ValueDefault { return nil }

func NewAny() interface{} { return nil }

var _ error = Interface(nil)

var _ error = (*struct{ error })(nil)

func Return2() error {
	var err error

	switch rand.Int() {
	case 0:
		err = NewInterface() // want "NOT IN RESULTS"

	case 1:
		err = NewStruct() // want "NOT IN RESULTS"

	case 2:
		err = NewValueDefault() // want "VALUE"

	case 3:
		err = &c.ValueDefault{} // want "VALUE"

	default:
		err = nil
	}

	return err
}
