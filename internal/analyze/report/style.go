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

// CheckStyle reports a diagnostic if the target of an errors.As-style function is not an address operation on a variable, suggesting a proper syntax.
func (r ErrorsAs) CheckStyle(tn types.Type) {
	if _, ok := r.varID(); ok {
		return
	}

	fname := r.funName()

	pkg := r.Pkg
	qf := func(other *types.Package) string {
		if other == pkg {
			return ""
		}

		return other.Name()
	}

	tname := types.TypeString(tn, qf)

	r.ReportRangef(r.Expr, `Target is not an address operation on a variable, use "var target %s; ... %s(err, &target)" instead. (et:sty)`,
		tname, fname)
}
