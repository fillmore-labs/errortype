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

import (
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Base provides the basic fields for reporting diagnostics.
// It holds the analysis pass and the AST node where the diagnostic should be reported.
type Base struct {
	*analysis.Pass
	Expr ast.Expr
}

// UndeterminedUsage reports a diagnostic for an error type with undetermined usage.
func (b Base) UndeterminedUsage(tn *types.TypeName, ptr bool) {
	fullName := types.TypeString(tn.Type(), nil)

	codeSuffix := ""
	if ptr {
		codeSuffix = "+"
	}

	b.ReportRangef(b.Expr,
		"Undetermined usage for error type %q. Specify in the configuration whether it is a pointer or value error. (et:emb%s)",
		fullName, codeSuffix)
}

func (b Base) relativeNameOf(tn *types.TypeName) string {
	return types.TypeString(tn.Type(), types.RelativeTo(b.Pkg))
}

func (b Base) shortNameOf(tn *types.TypeName) string {
	current := b.Pkg

	return types.TypeString(tn.Type(), func(pkg *types.Package) string {
		if pkg == current {
			return ""
		}

		return pkg.Name()
	})
}

var rawfmt = &printer.Config{Mode: printer.RawFormat}

func exprToString(fset *token.FileSet, expr ast.Expr) string {
	if sb := (strings.Builder{}); rawfmt.Fprint(&sb, fset, expr) == nil {
		return sb.String()
	}

	return "<invalid>"
}
