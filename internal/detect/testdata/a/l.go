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

//go:build go1.25

package a

import (
	"math/rand/v2"
	"reflect"
)

type (
	PointerCast struct{ error }
	ValueCast   struct{ error }
)

func IsPointerCast(err error) bool {
	v := reflect.ValueOf(err)
	_, ok := reflect.TypeAssert[*PointerCast](v)

	return ok
}

func IsValueCast(err error) bool {
	v := reflect.ValueOf(err)
	_, ok := reflect.TypeAssert[ValueCast](v)

	return ok
}

func ReturnL() error {
	var err error

	switch rand.Int() {
	case 0:
		err = PointerCast{} // want "POINTER"

	case 1:
		err = &ValueCast{} // want "VALUE"

	default:
		err = nil
	}

	return err
}
