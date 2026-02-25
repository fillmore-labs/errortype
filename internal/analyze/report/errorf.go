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

package report

import (
	"go/types"
)

// Errorf reports diagnostics related to operands wrapped by fmt.Errorf like functions.
type Errorf struct {
	Base
}

// ShouldBeValue reports a diagnostic when a value error is assigned as a pointer.
func (r Errorf) ShouldBeValue(tn *types.TypeName) {
	relativeName, shortName := r.relativeNameOf(tn), r.shortNameOf(tn)
	// fmt.Errorf("%w", &MyValueError{})
	r.ReportRangef(r.Expr,
		`Error type %q should be wrapped as a value ("%s{...}"), not a pointer. (et:wrp)`, relativeName, shortName)
}

// ShouldBePointer reports a diagnostic when a pointer error is assigned as a value.
func (r Errorf) ShouldBePointer(tn *types.TypeName) {
	relativeName, shortName := r.relativeNameOf(tn), r.shortNameOf(tn)
	// fmt.Errorf("%w", MyPointerError{})
	r.ReportRangef(r.Expr,
		`Error type %q should be wrapped as a pointer ("&%s{...}"), not a value. (et:wrp+)`, relativeName, shortName)
}
