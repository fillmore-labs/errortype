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

type (
	AliasError        = c.AliasError  // want AliasError:"value"
	PointerAliasError = *c.AliasError // want PointerAliasError:"value"
)

func Return() error {
	var err error

	switch rand.Int() {
	case 0:
		err = &c.ValueDefaultError{} // want "VALUE"

	case 1:
		err = &c.ValueFuncError{} // want "VALUE"

	case 2:
		err = &c.ValueVarError{} // want "VALUE"

	case 3:
		err = &c.PointerDefaultError{} // want "POINTER"

	case 4:
		err = c.PointerFuncError{} // want "POINTER"

	case 5:
		err = c.PointerVarError{} // want "POINTER"

	case 6:
		err = c.EmbeddedDefaultError{} // want "UNDECIDED"

	case 7:
		err = c.EmbeddedFuncError{} // want "POINTER"

	case 8:
		err = &c.EmbeddedVarError{} // want "VALUE"

	case 9:
		err = &c.AliasError{} // want "VALUE"

	case 10:
		err = c.PointerAliasError(&c.PointerDefaultError{}) // want "VALUE"

	case 11:
		err = (*AliasError)(&c.AliasError{}) // want "VALUE"

	case 12:
		err = PointerAliasError(&c.PointerDefaultError{}) // want "VALUE"

	default:
		err = nil
	}

	return err
}
