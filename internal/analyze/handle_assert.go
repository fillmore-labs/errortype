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

	"fillmore-labs.com/errortype/internal/typeutil"
)

// handleTypeAssert checks for incorrect pointer/value usage of error types in type assertions.
func (p pass) handleTypeAssert(n *ast.TypeAssertExpr) {
	if n.Type == nil {
		return // Type switches are handled in handleTypeSwitch
	}

	if tv, ok := p.TypesInfo.Types[n.X]; !ok || !typeutil.HasErrorMethod(tv.Type) {
		return // We are only interested in assertions on error interfaces
	}

	tv := p.TypesInfo.Types[n.Type]
	if !tv.IsType() { // should not happen
		p.ReportErrorf(n.Type, "Expected type, got %#v", tv)
	}

	p.checkErrorUsage(tv.Type, p.AssertReporter(n.Type))
}
