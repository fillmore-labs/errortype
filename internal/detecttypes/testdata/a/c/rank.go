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

package c

// Conversions to error interfaces outrank construction evidence.
type (
	RankLiteralError struct{ error } // want RankLiteralError:"value"
	RankCastError    int             // want RankCastError:"pointer"
)

func (RankCastError) Error() string { return "" }

func _(errs chan<- error) {
	e := &RankLiteralError{} // pointer literal
	errs <- *e               // value conversion
}

func _() {
	var err error

	_ = RankCastError(1)     // value cast
	err = new(RankCastError) // pointer conversion

	_ = err
}
