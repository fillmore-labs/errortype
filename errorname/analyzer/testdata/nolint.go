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

package testdata

import "errors"

//nolint:errorname
type Suppressed1Err struct{}

func (Suppressed1Err) Error() string { return "suppressed" }

// Suppressed2Err is intentionally named.
//
//nolint:other,errorname
type Suppressed2Err struct{}

func (Suppressed2Err) Error() string { return "suppressed list" }

var suppressed3Err = errors.New("suppressed all") //nolint:errorname

//nolint:all
var suppressed4Err = errors.New("suppressed all")

// HiddenErr would normally be flagged, but errortype is unrelated.
//
//nolint:errortype
type HiddenErr struct{} // want ` suggestion: "HiddenError"$`

func (HiddenErr) Error() string { return "hidden" }
