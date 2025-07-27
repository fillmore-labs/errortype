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
	"fmt"
	"go/ast"
	"io"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Fprint outputs the syntax tree representation of the given AST node `n` to `w`.
// This can be useful for debugging purposes.
func (p pass) Fprint(w io.Writer, n ast.Node) error {
	return ast.Fprint(w, p.Fset, n, ast.NotNilFilter)
}

// ReportErrorf reports an internal ("should not happen") failure message.
func (p pass) ReportErrorf(n ast.Node, format string, args ...any) {
	var sb strings.Builder
	_, _ = sb.WriteString("Internal error: ")
	_, _ = fmt.Fprintf(&sb, format, args...)
	_, _ = sb.WriteString(". (et:xxx)\n")
	_ = p.Fprint(&sb, n)

	p.Report(analysis.Diagnostic{
		Pos:      n.Pos(),
		End:      n.End(),
		Category: "xxx",
		Message:  sb.String(),
	})
}
