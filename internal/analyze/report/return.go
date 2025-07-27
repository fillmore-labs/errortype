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

package report

import "go/types"

// Return reports diagnostics related to return statements.
type Return struct {
	Base
}

// ShouldBeValue reports a diagnostic when a value error is returned as a pointer.
func (r Return) ShouldBeValue(tn *types.TypeName) {
	fullName, importName := r.relativeNameOf(tn), r.importNameOf(tn)
	// This case handles returning a pointer to a value-error ("return &MyValueError{}")
	r.ReportRangef(r.Expr,
		"Error type %q should be returned by value (\"%s{...}\"), not as a pointer. (et:ret)", fullName, importName)
}

// ShouldBePointer reports a diagnostic when a pointer error is returned as a value.
func (r Return) ShouldBePointer(tn *types.TypeName) {
	fullName, importName := r.relativeNameOf(tn), r.importNameOf(tn)
	// This case handles returning a value of a pointer-error ("return MyPointerError{}")
	r.ReportRangef(r.Expr,
		"Error type %q should be returned as a pointer (\"&%s{...}\"), not by value. (et:ret+)", fullName, importName)
}
