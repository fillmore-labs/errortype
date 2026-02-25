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

type my2Error struct{}

type alias2Error = *my2Error

func (alias2Error) Error() string {
	return "alias2Error"
}

func (alias2Error) String() string {
	return "alias2Error"
}

type alias3Error = *struct{ error }

func Alias() {
	var err interface {
		fmt.Stringer
		Error() string
	} = &my2Error{}

	_, _ = err.(alias2Error)

	_, _ = err.(*my2Error)

	var e2 *my2Error

	_ = errors.As(err, &e2)

	var a2 alias2Error

	_ = errors.As(err, &a2)

	switch err.(type) {
	case *my2Error:
	}

	switch err.(type) {
	case alias2Error:
	}

	var err3 error = alias3Error(&struct{ error }{error: err})

	var e3 alias3Error
	_ = errors.As(err3, e3) // want ` \(et:sty\)$`

	_ = errors.As(err3, &e3)
}
