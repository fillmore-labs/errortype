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

type Alias = c.Alias

func Return() error {
	switch rand.Int() {
	case 0:
		return &c.ValueDefault{} // want "VALUE"

	case 1:
		return &c.ValueFunc{} // want "VALUE"

	case 2:
		return &c.ValueVar{} // want "VALUE"

	case 3:
		return &c.PointerDefault{} // want "POINTER"

	case 4:
		return &c.PointerFunc{} // want "POINTER"

	case 5:
		return &c.PointerVar{} // want "POINTER"

	case 6:
		return &c.EmbeddedDefault{}

	case 7:
		return &c.EmbeddedFunc{} // want "POINTER"

	case 8:
		return &c.EmbeddedVar{} // want "VALUE"

	case 9:
		return &c.Alias{} // want "VALUE"

	default:
		return nil
	}
}
