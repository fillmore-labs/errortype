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

// ErrorsAs reports diagnostics related to targets in errors.As like functions.
type ErrorsAs struct {
	Base
	Fun *types.Func
}

// ShouldBeValue reports a diagnostic for a mismatch between expected and actual error usage in a function call.
func (r ErrorsAs) ShouldBeValue(tn *types.TypeName) {
	fullName, importName := r.relativeNameOf(tn), r.importNameOf(tn)
	fname, varname := r.funName(), r.varName()

	// errors.As(err, &p) where p is *ValueError. Target is **ValueError, but should be *ValueError.
	r.ReportRangef(r.Expr, `Target for value error %q is a pointer-to-pointer, use a pointer to a value instead: "var %s %s; ... %s(err, &%s)". (et:err)`,
		fullName, varname, importName, fname, varname)
}

// ShouldBePointer reports a diagnostic for a mismatch between expected and actual error usage in a function call.
func (r ErrorsAs) ShouldBePointer(tn *types.TypeName) {
	fullName, importName := r.relativeNameOf(tn), r.importNameOf(tn)
	fname, varname := r.funName(), r.varName()

	// errors.As(err, &p) where p is PointerError. Target is *PointerError, but should be **PointerError.
	r.ReportRangef(r.Expr, `Target for pointer error %q is a pointer-to-value, use a pointer to a pointer instead: "var %s *%s; ... %s(err, &%s)". (et:err+)`,
		fullName, varname, importName, fname, varname)
}

// funName gets a short function name, not necessarily matching imports.
func (r ErrorsAs) funName() string {
	if pkg := r.Fun.Pkg(); pkg != nil {
		return pkg.Name() + "." + r.Fun.Name()
	}

	return r.Fun.Name()
}
