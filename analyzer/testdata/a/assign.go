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

package a

import (
	"net"
	"strconv"

	"test/a/d"
)

type IntegerError int

func (i IntegerError) Error() string { return "error " + strconv.Itoa(int(i)) }

func Assign() (err net.Error) {
	a, err := IntegerError(1), &d.ValueError{} // want ` \(et:asn\)$`

	err = d.PointerError{} // want ` \(et:asn\+\)$`

	a += 1

	return
}

func AssignSimple() {
	var err error

	err = new(IntegerError) // want ` \(et:asn\)$`

	_ = err
}
