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

// Switch reports diagnostics related to type assertions in switch cases.
type Switch struct {
	Base
}

// ShouldBeValue reports a diagnostic when a value error is asserted as a pointer.
func (r Switch) ShouldBeValue(tn *types.TypeName) {
	fullName, importName := r.relativeNameOf(tn), r.importNameOf(tn)
	// "_, ok := err.(*MyValueError)" or "case *MyValueError:"
	r.ReportRangef(r.Expr,
		`Value error %q should be used as a value type ("case %s:") in the type switch, not as a pointer type. (et:ast)`, fullName, importName)
}

// ShouldBePointer reports a diagnostic when a pointer error is asserted as a value.
func (r Switch) ShouldBePointer(tn *types.TypeName) {
	fullName, importName := r.relativeNameOf(tn), r.importNameOf(tn)
	// "_, ok := err.(MyPointerError)"" or "case MyPointerError:""
	r.ReportRangef(r.Expr,
		`Pointer error %q should be used as a pointer type ("case *%s:") in the type switch, not as a value type. (et:ast+)`, fullName, importName)
}
