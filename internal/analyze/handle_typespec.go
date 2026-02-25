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

package analyze

import (
	"go/ast"
	"go/types"

	"fillmore-labs.com/errortype/internal/astutil"
)

// handleTypeSpec checks for missing Is or Unwrap methods on not comparable value error types.
func (p Pass) handleTypeSpec(spec *ast.TypeSpec, decl *ast.GenDecl) {
	id := spec.Name
	if id.Name == "_" {
		return
	}

	tn, ok := p.TypesInfo.Defs[id].(*types.TypeName)
	if !ok { // should not happen
		return
	}

	if !p.IsValueError(tn) || types.Comparable(tn.Type()) {
		return
	}

	if _, ok := astutil.CheckTypeSpecComments(spec, decl, p.Analyzer.Name); !ok {
		return
	}

	codeSuffix := ""
	if tn.Exported() {
		codeSuffix = "+"
	}

	p.ReportRangef(spec, "Not comparable error type %q should be a pointer type. (et:nce%s)", tn.Name(), codeSuffix)
}
