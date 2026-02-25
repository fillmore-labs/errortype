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
	Alias        = c.Alias  // want Alias:"value"
	PointerAlias = *c.Alias // want PointerAlias:"value"
)

func Return() error {
	var err error

	switch rand.Int() {
	case 0:
		err = &c.ValueDefault{} // want "VALUE"

	case 1:
		err = &c.ValueFunc{} // want "VALUE"

	case 2:
		err = &c.ValueVar{} // want "VALUE"

	case 3:
		err = &c.PointerDefault{} // want "POINTER"

	case 4:
		err = c.PointerFunc{} // want "POINTER"

	case 5:
		err = c.PointerVar{} // want "POINTER"

	case 6:
		err = c.EmbeddedDefault{} // want "UNDECIDED"

	case 7:
		err = c.EmbeddedFunc{} // want "POINTER"

	case 8:
		err = &c.EmbeddedVar{} // want "VALUE"

	case 9:
		err = &c.Alias{} // want "VALUE"

	case 10:
		err = c.PointerAlias(&c.PointerDefault{}) // want "VALUE"

	case 11:
		err = (*Alias)(&c.Alias{}) // want "VALUE"

	case 12:
		err = PointerAlias(&c.PointerDefault{}) // want "VALUE"

	default:
		err = nil
	}

	return err
}
