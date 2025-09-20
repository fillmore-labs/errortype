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

package detect

import (
	"fmt"
	"go/ast"
	"go/types"
	"log"
	"strings"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// LogErrorf reports an internal ("should not happen") failure message.
func (p pass) LogErrorf(n ast.Node, format string, args ...any) {
	var sb strings.Builder
	_, _ = sb.WriteString("Internal error: ")
	_, _ = fmt.Fprintf(&sb, format, args...)
	_ = sb.WriteByte('\n')

	_ = ast.Fprint(&sb, p.Fset, n, ast.NotNilFilter)

	log.Println(sb.String())
}

// logResults logs the determined error types for each type name in the PropertyMap.
func (p pass) logResults() {
	name := typeutil.PkgName(p.Pass)

	qf := types.RelativeTo(p.Pkg)

	for tn, errortype := range p.AllSorted {
		typeName := types.TypeString(tn.Type(), qf)
		determinedType := errortype.DeterminedType()

		log.Printf("%s %s: %s (%s)", name, typeName, determinedType, errortype)
	}
}
