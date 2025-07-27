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

import "golang.org/x/tools/go/ast/inspector"

// handleReturns identifies function return parameters that are of type error
// and then inspects all return statements within the function body to check
// for incorrect error type usage.
func (p pass) handleReturns(b inspector.Cursor, lastResult int) {
	for retStmt := range AllReturns(b) {
		if len(retStmt.Results) <= lastResult {
			continue // Skip return statements with differing arity
		}

		res := retStmt.Results[lastResult]

		resType := p.TypesInfo.Types[res]
		if !resType.IsValue() { // should not happen
			p.ReportErrorf(res, "Expected value, got %#v", resType)
		}

		if resType.IsNil() {
			continue // nil is fine.
		}

		p.checkErrorUsage(resType.Type, p.ReturnReporter(res))
	}
}
