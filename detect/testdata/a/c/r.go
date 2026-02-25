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

package c

import "errors"

type (
	AmbiguousPointerError   struct{ error }                 // want AmbiguousPointerError:"pointer"
	AmbiguousValueError     struct{ error }                 // want AmbiguousValueError:"value"
	AmbiguousAmbiguousError struct{ error }                 // want AmbiguousAmbiguousError:"undecided"
	EmbeddedPointerError    struct{ PointerDefaultError }   // want EmbeddedPointerError:"pointer"
	EmbeddedValueError      struct{ ValueDefaultError }     // want EmbeddedValueError:"undecided"
	EmbeddedAmbiguousError  struct{ EmbeddedDefaultError }  // want EmbeddedAmbiguousError:"undecided"
	EmbeddedPPointerError   struct{ *PointerDefaultError }  // want EmbeddedPPointerError:"undecided"
	EmbeddedPValueError     struct{ *ValueDefaultError }    // want EmbeddedPValueError:"undecided"
	EmbeddedPAmbiguousError struct{ *EmbeddedDefaultError } // want EmbeddedPAmbiguousError:"undecided"
)

func (a *AmbiguousPointerError) As(target any) bool { return errors.As(a.error, target) }
func (a AmbiguousValueError) Unwrap() error         { return a.error }

func (a AmbiguousAmbiguousError) As() error               { return a.error }
func (a *AmbiguousAmbiguousError) Unwrap(target any) bool { return errors.As(a.error, target) }
