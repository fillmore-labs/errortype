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

//go:build go1.26

package a

import (
	"errors"
	"math/rand/v2"
)

type AsTypeCalledError struct{ error } // want AsTypeCalledError:"pointer"

func ReturnL126(err error) error {
	_, _ = errors.AsType[*AsTypeCalledError](err)

	switch rand.Int() {
	case 0:
		err = AsTypeCalledError{} // want "POINTER"

	default:
		err = nil
	}

	return err
}
