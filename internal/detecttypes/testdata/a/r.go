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

func Return4() error {
	var err error

	switch rand.Int() {
	case 0:
		err = c.AmbiguousPointerError{} // want "POINTER"

	case 1:
		err = &c.AmbiguousValueError{} // want "VALUE"

	case 2:
		err = c.AmbiguousAmbiguousError{} // want "UNDECIDED"

	case 3:
		err = &c.EmbeddedPointerError{} // want "POINTER"

	case 4:
		err = c.EmbeddedValueError{} // want "UNDECIDED"

	case 5:
		err = c.EmbeddedAmbiguousError{} // want "UNDECIDED"

	case 6:
		err = c.EmbeddedPPointerError{} // want "UNDECIDED"

	case 7:
		err = c.EmbeddedPValueError{} // want "UNDECIDED"

	case 8:
		err = c.EmbeddedPAmbiguousError{} // want "UNDECIDED"

	default:
		err = nil
	}

	return err
}
