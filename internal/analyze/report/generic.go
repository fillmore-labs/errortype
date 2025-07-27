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

// Generic reports diagnostics related to generic function calls.
type Generic struct {
	Base
	Fun *types.Func
}

// ShouldBeValue reports a diagnostic when a value error is queried as a pointer.
func (r Generic) ShouldBeValue(tn *types.TypeName) {
	fullName, importName := r.relativeNameOf(tn), r.importNameOf(tn)
	fname := r.funName()

	r.ReportRangef(r.Expr,
		`Error type %q should be queried as a value ("%s[%s]"), not a pointer. (et:ast)`, fullName, fname, importName)
}

// ShouldBePointer reports a diagnostic when a pointer error is queried as a value.
func (r Generic) ShouldBePointer(tn *types.TypeName) {
	fullName, importName := r.relativeNameOf(tn), r.importNameOf(tn)
	fname := r.funName()

	r.ReportRangef(r.Expr,
		`Error type %q should be queried as a pointer ("%s[*%s]"), not a value. (et:ast+)`, fullName, fname, importName)
}

// funName gets a short function name, not necessarily matching imports.
func (r Generic) funName() string {
	if pkg := r.Fun.Pkg(); pkg != nil {
		return pkg.Name() + "." + r.Fun.Name()
	}

	return r.Fun.Name()
}
