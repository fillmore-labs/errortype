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

// VarDecl reports diagnostics related to variable declarations.
type VarDecl struct {
	Base
	VarName string
}

// ShouldBeValue reports a diagnostic when a value error is assigned as a pointer.
func (r VarDecl) ShouldBeValue(tn *types.TypeName) {
	if relativeName := r.relativeNameOf(tn); r.VarName == "_" {
		r.ReportRangef(r.Expr,
			`Compile‑time assertion should be a value ("... = %s{...}"), not a pointer. (et:var)`, relativeName)
	} else {
		r.ReportRangef(r.Expr,
			`Error %q should be a value ("... = %s{...}"), not a pointer. (et:var)`, r.VarName, relativeName)
	}
}

// ShouldBePointer reports a diagnostic when a pointer error is assigned as a value.
func (r VarDecl) ShouldBePointer(tn *types.TypeName) {
	if relativeName := r.relativeNameOf(tn); r.VarName == "_" {
		r.ReportRangef(r.Expr,
			`Compile‑time assertion should be a pointer ("... = &%s{...}"), not a value. (et:var+)`, relativeName)
	} else {
		r.ReportRangef(r.Expr,
			`Error %q should be a pointer ("... = &%s{...}"), not a value. (et:var+)`, r.VarName, relativeName)
	}
}
