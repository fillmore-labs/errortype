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

//go:build go1.23

package testdata

import (
	"fmt"
	"time"
)

// ErrorOldGeneric[T] is an exported generic error.
// It should not be renamed, since generial aliases require Go 1.24.
type ErrorOldGeneric[T fmt.Stringer] struct{ v T } // want ` suggestion: "OldGenericError"$`

func (e ErrorOldGeneric[T]) Error() string { return "error: " + e.v.String() }

// errorOldGeneric[T] is an unexported generic error.
// It should  be renamed, since it does not require a deprecation alias.
type errorOldGeneric[T fmt.Stringer] struct{ v T } // want ` suggestion: "oldGenericError"$`

func (e errorOldGeneric[T]) Error() string { return "error: " + e.v.String() }

// ErrorOld is old.
type ErrorOld = ErrorOldGeneric[time.Time] // want ` suggestion: "OldError"$`

// errorOld is old, too.
type errorOld = errorOldGeneric[time.Time] // want ` suggestion: "oldError"$`
