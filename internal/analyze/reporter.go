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

package analyze

import (
	"go/ast"
	"go/types"

	"fillmore-labs.com/errortype/internal/analyze/report"
)

// UsageReporter defines the interface for reporting diagnostics related to
// incorrect error type usage. Different implementations can provide
// context-specific error messages for returns or type assertions.
type UsageReporter interface {
	// ShouldBeValue is called when an error type that should be a value is
	// used as a pointer.
	ShouldBeValue(tn *types.TypeName)

	// ShouldBePointer is called when an error type that should be a pointer is
	// used as a value.
	ShouldBePointer(tn *types.TypeName)

	// UndeterminedUsage is called when a named error type is encountered whose
	// pointer-vs-value usage has not been determined by the `detecttypes`
	// analyzer. This is often due to embedding the `error` interface.
	UndeterminedUsage(tn *types.TypeName, isPtr bool)
}

// AssertReporter creates a new reporter for assertions.
func (p pass) AssertReporter(e ast.Expr) report.Assert {
	return report.Assert{Base: report.Base{Pass: p.Pass, Expr: e}}
}

// ErrorsAsReporter creates a new reporter for errors.As like functions.
func (p pass) ErrorsAsReporter(e ast.Expr, fun *types.Func) report.ErrorsAs {
	return report.ErrorsAs{Base: report.Base{Pass: p.Pass, Expr: e}, Fun: fun}
}

// ReturnReporter creates a new reporter for return statements.
func (p pass) ReturnReporter(e ast.Expr) report.Return {
	return report.Return{Base: report.Base{Pass: p.Pass, Expr: e}}
}

// SwitchReporter creates a new reporter for type switches.
func (p pass) SwitchReporter(e ast.Expr) report.Switch {
	return report.Switch{Base: report.Base{Pass: p.Pass, Expr: e}}
}

// GenericReporter creates a new reporter for generic functions.
func (p pass) GenericReporter(e ast.Expr, fun *types.Func) report.Generic {
	return report.Generic{Base: report.Base{Pass: p.Pass, Expr: e}, Fun: fun}
}
