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
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// Base provides the basic fields for reporting diagnostics.
// It holds the analysis pass and the AST node where the diagnostic should be reported.
type Base struct {
	*analysis.Pass
	Expr ast.Expr
}

// UndeterminedUsage reports a diagnostic for an error type with undetermined usage.
func (r Base) UndeterminedUsage(tn *types.TypeName, isPtr bool) {
	fullName := r.relativeNameOf(tn)

	plus := ""
	if isPtr {
		plus = "+"
	}

	r.ReportRangef(r.Expr,
		"Undetermined usage for error type %q. Specify in the configuration whether it is a pointer or value error. (et:emb%s)",
		fullName, plus)
}

func (r Base) relativeNameOf(tn *types.TypeName) string {
	return types.TypeString(tn.Type(), types.RelativeTo(r.Pkg))
}

func (r Base) importNameOf(tn *types.TypeName) string {
	current := r.Pkg

	return types.TypeString(tn.Type(), func(pkg *types.Package) string {
		if pkg == current {
			return ""
		}

		return pkg.Name()
	})
}

// varName gets the target variable name when it is the expression "&name", a generic "target" otherwise.
func (r Base) varName() string {
	if id, ok := r.varID(); ok {
		return id.Name
	}

	return "target"
}

func (r Base) varID() (*ast.Ident, bool) {
	if e, ok := ast.Unparen(r.Expr).(*ast.UnaryExpr); ok && e.Op == token.AND {
		id, ok := ast.Unparen(e.X).(*ast.Ident)

		return id, ok
	}

	return nil, false
}
