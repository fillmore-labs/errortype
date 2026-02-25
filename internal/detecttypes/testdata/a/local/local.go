// Copyright 2026 Oliver Eikemeier. All Rights Reserved.
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

package local

import (
	"math/rand/v2"

	"test/a/c"
)

// Plain tie-breaker on imported undecided type, plus plain disagreeing assertion (ignored).
var (
	_ error = (*c.EmbeddedDefaultError)(nil)
	_ error = (*c.ValueDefaultError)(nil)
)

// Directive overrides on imported decided types and mismatched directive (ignored).
var (
	//errortype:pointer
	_ error = (*c.ValueFuncError)(nil)

	//errortype:value
	_ error = c.PointerFuncError{}

	//errortype:suppress
	_ error = (*c.ValueVarError)(nil)

	//errortype:suppress
	_ error = c.PointerVarError{}

	//errortype:pointer
	_ error = c.EmbeddedVarError{}
)

func ReturnLocalOverrides() error {
	var err error

	switch rand.Int() {
	case 0:
		err = c.EmbeddedDefaultError{} // want "POINTER"

	case 1:
		err = &c.ValueDefaultError{} // want "VALUE"

	case 2:
		err = c.ValueFuncError{} // want "POINTER"

	case 3:
		err = &c.PointerFuncError{} // want "VALUE"

	case 4:
		err = c.ValueVarError{} // want "SUPPRESS"

	case 5:
		err = &c.PointerVarError{} // want "SUPPRESS"

	case 6:
		err = &c.EmbeddedVarError{} // want "VALUE"

	default:
		err = nil
	}

	return err
}
