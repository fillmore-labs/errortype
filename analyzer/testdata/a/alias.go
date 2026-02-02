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
	"errors"
	"fmt"
)

type myError2 struct{}

type aliasError2 = *myError2

func (aliasError2) Error() string {
	return "aliasError2"
}

func (aliasError2) String() string {
	return "aliasError2"
}

type aliasError3 = *struct{ error }

func Alias() {
	var err interface {
		fmt.Stringer
		Error() string
	} = &myError2{}

	_, _ = err.(aliasError2)

	_, _ = err.(*myError2)

	var e2 *myError2

	_ = errors.As(err, &e2)

	var a2 aliasError2

	_ = errors.As(err, &a2)

	switch err.(type) {
	case *myError2:
	}

	switch err.(type) {
	case aliasError2:
	}

	var err3 error = aliasError3(&struct{ error }{error: err})

	var e3 aliasError3
	_ = errors.As(err3, e3) // want " \\(et:sty\\)$"

	_ = errors.As(err3, &e3)
}
