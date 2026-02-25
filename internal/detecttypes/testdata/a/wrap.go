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

// Wrapped by %w as a value, queried as a pointer: wrapping decides, like a return.
type WrapValueError struct{ msg string } // want WrapValueError:"value"

func (e WrapValueError) Error() string { return e.msg }

// Wrapped by %w as a pointer, queried as a value.
type WrapPointerError struct{ msg string } // want WrapPointerError:"pointer"

func (e WrapPointerError) Error() string { return e.msg }

// A wrapped variable counts, not only a composite literal.
type WrapVarError struct{ msg string } // want WrapVarError:"value"

func (e WrapVarError) Error() string { return e.msg }

// %v does not wrap, so the query decides.
type FormatVerbError struct{ msg string } // want FormatVerbError:"pointer"

func (e FormatVerbError) Error() string { return e.msg }

// A pointer Unwrap method still decides over wrapping.
type WrapUnwrapError struct{ err error } // want WrapUnwrapError:"pointer"

func (e WrapUnwrapError) Error() string { return e.err.Error() }

func (e *WrapUnwrapError) Unwrap() error { return e.err }

// Wrapped by a fmt.Errorf-like wrapper.
type WrapperWrapError struct{ msg string } // want WrapperWrapError:"value"

func (e WrapperWrapError) Error() string { return e.msg }

func wrapf(format string, args ...any) error { return fmt.Errorf(format, args...) }

func Wrap(err error, e WrapVarError) error {
	_ = errors.As(err, new(*WrapValueError))
	_ = errors.As(err, new(WrapPointerError))
	_, _ = errors.AsType[*WrapVarError](err)
	_, _ = errors.AsType[*FormatVerbError](err)
	_, _ = errors.AsType[*WrapperWrapError](err)

	_ = fmt.Errorf("value: %w", WrapValueError{})
	_ = fmt.Errorf("pointer: %w", &WrapPointerError{})
	_ = fmt.Errorf("variable: %w", e)
	_ = fmt.Errorf("format: %v", FormatVerbError{})
	_ = wrapf("wrapper: %w", WrapperWrapError{})

	return fmt.Errorf("unwrap: %w", WrapUnwrapError{err})
}
