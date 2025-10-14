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
	"strings"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// handleVarDecls checks for incorrect pointer/value usage of error types in variable declarations.
func (p Pass) handleVarDecls(n *ast.ValueSpec) {
	for i, id := range n.Names {
		if len(n.Values) <= i {
			break
		}

		if id.Name != "_" && !strings.HasPrefix(id.Name, "Err") && !strings.HasPrefix(id.Name, "err") {
			continue
		}

		value := n.Values[i]

		tv, ok := p.TypesInfo.Types[value]
		if !ok || !typeutil.HasErrorMethod(tv.Type) {
			continue // Not an error type
		}

		p.checkVarDecl(tv.Type, value, id.Name)
	}
}
