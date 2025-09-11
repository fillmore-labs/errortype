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

import "math/rand/v2"

type PointerOverride struct{}

func (PointerOverride) Error() string { return "" }

type ValueOverride struct{}

func (ValueOverride) Error() string { return "" }

type SuppressOverride struct{ error }

var (
	_ error = PointerOverride{}
	_ error = &ValueOverride{}
)

func ReturnOverride() error {
	switch rand.Int() {
	case 0:
		return &PointerOverride{}

	case 1:
		return ValueOverride{}

	case 2:
		return SuppressOverride{}

	case 3:
		return &SuppressOverride{}

	default:
		return nil
	}
}
