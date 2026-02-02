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

package c

import (
	"errors"
	"strconv"
)

type myIntError int

func (e myIntError) Error() string { return "error " + strconv.Itoa(int(e)) }

var _ error = myIntError(0)

func Basic(err *myIntError) {
	_ = new(int) == new((int)) // want " \\(et:equ\\+\\)"

	errors.Is(err, new(myIntError)) // want " \\(et:cmp\\+\\)" " \\(et:unu\\+\\)"
}
