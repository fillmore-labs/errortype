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

// Send reports diagnostics related to error values sent on channels.
type Send struct {
	Base
}

// ShouldBeValue reports a diagnostic when a value error is sent on a channel as a pointer.
func (r Send) ShouldBeValue(tn *types.TypeName) {
	relativeName, form := r.relativeNameOf(tn), r.valueForm(tn)
	// ch <- &MyValueError{}
	r.ReportRangef(r.Expr,
		`Error type %q should be sent as a value ("%s"), not a pointer. (et:asn)`, relativeName, form)
}

// ShouldBePointer reports a diagnostic when a pointer error is sent on a channel as a value.
func (r Send) ShouldBePointer(tn *types.TypeName) {
	relativeName, form := r.relativeNameOf(tn), r.pointerForm(tn)
	// ch <- MyPointerError{}
	r.ReportRangef(r.Expr,
		`Error type %q should be sent as a pointer ("%s"), not a value. (et:asn+)`, relativeName, form)
}
