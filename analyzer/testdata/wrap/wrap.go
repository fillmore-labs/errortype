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

package wrap

import (
	"errors"
	"fmt"
)

// ValueError is wrapped by %w as a value, which decides its usage like a return.
type ValueError struct{ msg string }

func (e ValueError) Error() string { return e.msg }

func Wrap() error {
	return fmt.Errorf("wrap: %w", ValueError{"value"})
}

// Queries in the pointer form are reported, not the wrapping.

func QueryAsType(err error) bool {
	_, ok := errors.AsType[*ValueError](err) // want ` \(et:ast\)$`
	return ok
}

func QueryAs(err error) bool {
	var target *ValueError
	return errors.As(err, &target) // want ` \(et:err\)$`
}

func QueryValue(err error) bool {
	_, ok := errors.AsType[ValueError](err)
	return ok
}
