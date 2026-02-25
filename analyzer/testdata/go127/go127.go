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

//go:build go1.27

package go127

import (
	"errors"
	"fmt"
)

// In files at Go 1.27, a value receiver on Error of a pointer error is reported by default,
// since go vet reports wrapping the pointer with %w.
type StatusError struct{ Code int }

func (e StatusError) Error() string { return fmt.Sprintf("status %d", e.Code) } // want ` \(et:rcv\+\)$`

func Get() error { return fmt.Errorf("get: %w", &StatusError{404}) }

func Code(err error) int {
	if serr, ok := errors.AsType[*StatusError](err); ok {
		return serr.Code
	}

	return 0
}

// A promoted Error is not reported, with or without pointer methods, even though go vet reports wrapping the pointer.
type PromotedError struct{ error }

func GetPromoted(err error) error { return fmt.Errorf("get: %w", &PromotedError{err}) }

type PromotedUnwrapError struct{ error }

func (e *PromotedUnwrapError) Unwrap() error { return e.error }

func GetPromotedUnwrap(err error) error { return fmt.Errorf("get: %w", &PromotedUnwrapError{err}) }
