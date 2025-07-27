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

package main

import (
	"errors"
	"fmt"
)

type myError struct{ _ int }

func (*myError) Error() string { return "" }

type (
	aliasError  = *struct{ myError }
	aliasError2 = struct{ *myError }
	aliasError3 = struct{ myError }
)

var _, _, _ error = aliasError(nil), aliasError2(struct{ *myError }{}), func() *aliasError3 { return nil }()

func main() {
	var err error = aliasError(&struct{ myError }{myError{}})

	var e aliasError
	if errors.As(err, &e) {
		fmt.Println("&e")
	}

	var err2 error = aliasError2(struct{ *myError }{&myError{}})

	var e2 aliasError2
	if errors.As(err2, &e2) {
		fmt.Println("&e2")
	}

	a3 := aliasError3(struct{ myError }{myError{}})
	var err3 error = &a3

	var ep3 *aliasError3
	if errors.As(err3, &ep3) {
		fmt.Println("&ep3")
	}
}
