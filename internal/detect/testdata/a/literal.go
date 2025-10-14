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
)

type (
	PointerLiteral1 struct{ error }
	PointerLiteral2 struct{ error }
	PointerLiteral3 struct{ error }
	ValueLiteral    struct{ error }
)

func ReturnLiteral(err error) error {
	err1 := []*PointerLiteral1{
		{err},
	}

	err2 := [...]*PointerLiteral1{
		{err},
	}

	err3 := map[int]*PointerLiteral1{
		1: {err},
	}

	err4 := []ValueLiteral{
		{err},
	}

	switch rand.Int() {
	case 0:
		err = err1[0] // want "POINTER"

	case 1:
		err = err2[0] // want "POINTER"

	case 2:
		err = err3[0] // want "POINTER"

	case 3:
		err = err4[0] // want "VALUE"

	default:
		err = nil
	}

	return err
}
