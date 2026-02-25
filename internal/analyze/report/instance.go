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

import "go/types"

// Instance reports diagnostics related to generic instantiations of error-constrained type parameters.
type Instance struct {
	Base
}

// ShouldBeValue reports a diagnostic when a value error instantiates a type parameter as a pointer.
func (r Instance) ShouldBeValue(tn *types.TypeName) {
	relativeName, shortName := r.relativeNameOf(tn), r.shortNameOf(tn)
	r.ReportRangef(r.Expr,
		`Error type %q should be instantiated as a value ("%s"), not a pointer. (et:ins)`, relativeName, shortName)
}

// ShouldBePointer reports a diagnostic when a pointer error instantiates a type parameter as a value.
func (r Instance) ShouldBePointer(tn *types.TypeName) {
	relativeName, shortName := r.relativeNameOf(tn), r.shortNameOf(tn)
	r.ReportRangef(r.Expr,
		`Error type %q should be instantiated as a pointer ("*%s"), not a value. (et:ins+)`, relativeName, shortName)
}
