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

package receiversall

import (
	"errors"
	"fmt"
	"net"
)

// A value error type used as a pointer: go vet (go 1.27) reports wrapping it with %w.
type StatusError struct{ Code int }

func (e StatusError) Error() string { return fmt.Sprintf("status %d", e.Code) } // want ` \(et:rcv\+\)$`

func Get() error { return fmt.Errorf("get: %w", &StatusError{404}) }

// Mixed receivers keep their specific message.
type MixedError struct{ err error }

func (e MixedError) Error() string { return e.err.Error() } // want `like "Unwrap" \(et:rcv\+\)$`

func (e *MixedError) Unwrap() error { return e.err }

// Value receivers on other methods of a pointer error are fine: they are in the method set of the pointer.
type TimeoutError struct{ op string }

func (e *TimeoutError) Error() string { return e.op + ": timeout" }

func (e TimeoutError) Timeout() bool { return true }

func (e TimeoutError) Temporary() bool { return true }

var _ net.Error = (*TimeoutError)(nil)

// A value error with a pointer method: net.Error finds Timeout only on the pointer.
type DeadlineError struct{ op string }

func (e DeadlineError) Error() string { return e.op + ": deadline exceeded" }

func (e *DeadlineError) Timeout() bool { return true } // want ` \(et:rcv\)$`

func (e DeadlineError) Temporary() bool { return true }

var ErrDeadline = DeadlineError{"read"}

// Decoders and unexported methods of a value error may have pointer receivers.
type CodeError struct{ code int }

func (e CodeError) Error() string { return fmt.Sprintf("code %d", e.code) }

func (e CodeError) MarshalText() ([]byte, error) { return []byte(fmt.Sprint(e.code)), nil }

func (e *CodeError) UnmarshalText(text []byte) error {
	_, err := fmt.Sscan(string(text), &e.code)
	return err
}

func (e *CodeError) Scan(src any) error { _, err := fmt.Sscan(fmt.Sprint(src), &e.code); return err }

func (e *CodeError) Set(s string) error { return e.UnmarshalText([]byte(s)) }

func (e CodeError) String() string { return e.Error() }

func (e *CodeError) reset() { e.code = 0 }

func (e *CodeError) Code() int { return e.code } // want ` \(et:rcv\)$`

var ErrCode = CodeError{42}

// Consistent receivers are fine.
type ValueError struct{ msg string }

func (e ValueError) Error() string { return e.msg }

func (e ValueError) Timeout() bool { return false }

var ErrValue = ValueError{"value"}

type PointerError struct{ msg string }

func (e *PointerError) Error() string { return e.msg }

func (e *PointerError) Timeout() bool { return false }

var ErrPointer = &PointerError{"pointer"}

func Timeout(err error) bool {
	netErr, ok := errors.AsType[net.Error](err)

	return ok && netErr.Timeout()
}
