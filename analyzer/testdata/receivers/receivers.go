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

package receivers

import "errors"

// A value Error with a pointer Unwrap: the value implements error, but can't unwrap.
type MixedUnwrapError struct{ err error }

func (e MixedUnwrapError) Error() string { return e.err.Error() } // want ` \(et:rcv\+\)$`

func (e *MixedUnwrapError) Unwrap() error { return e.err }

// The same with Is and As.
type MixedIsError struct{ _ int }

func (MixedIsError) Error() string { return "is" } // want ` \(et:rcv\+\)$`

func (*MixedIsError) Is(error) bool { return false }

type MixedAsError struct{ _ int }

func (MixedAsError) Error() string { return "as" } // want ` \(et:rcv\+\)$`

func (*MixedAsError) As(any) bool { return false }

// Consistent pointer receivers are fine.
type PointerError struct{ err error }

func (e *PointerError) Error() string { return e.err.Error() }

func (e *PointerError) Unwrap() error { return e.err }

// Consistent value receivers are fine.
type ValueError struct{ err error }

func (e ValueError) Error() string { return e.err.Error() }

func (e ValueError) Unwrap() error { return e.err }

// A value Error on a type used as a pointer is fine without pointer Unwrap, Is or As methods.
type UsedAsPointerError struct{ _ int }

func (UsedAsPointerError) Error() string { return "used as pointer" }

var ErrUsedAsPointer error = &UsedAsPointerError{}

// The diagnostic can be suppressed.
type SuppressedError struct{ err error }

//nolint:errortype
func (e SuppressedError) Error() string { return e.err.Error() }

func (e *SuppressedError) Unwrap() error { return e.err }

var _ = errors.New
