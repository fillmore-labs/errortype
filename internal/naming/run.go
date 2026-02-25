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

package naming

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/astutil"
	"fillmore-labs.com/errortype/internal/naming/diagnostic"
	"fillmore-labs.com/errortype/internal/naming/rules"
)

// Run flows through two main stages:
//
//   - [rules] discovers non-conforming error declarations and records their
//     doc-comment metadata.
//   - [diagnostic] emits diagnostics and suggested fixes.
func (o *Options) Run(pass *analysis.Pass) (any, error) {
	fset, files := pass.Fset, pass.Files
	fileMap := astutil.NewFileMap(fset, files)
	linter := pass.Analyzer.Name

	violations := o.check(pass.TypesInfo, fileMap, files, linter)

	return nil, diagnostic.ReportViolations(pass, fileMap, violations, o.Message)
}

// check walks the definitions in files and returns one entry per
// error-typed constant, variable, or named type whose name does not follow
// the [error naming] convention. The result is sorted by declaration position.
//
// [error naming]: https://go.dev/wiki/Errors#naming
func (o *Options) check(info *types.Info, fileMap astutil.FileMap, files []*ast.File, linter string) []rules.Violation {
	checker := rules.New(info, linter)

	// Loop over all files
	for _, file := range files {
		if len(file.Decls) == 0 {
			continue
		}

		// Skip generated files
		if !o.Generated && fileMap.IsGenerated(file) {
			continue
		}

		// Skip files with a nolint comment
		if astutil.HasNoLint(file.Doc, linter) {
			continue
		}

		for _, decl := range file.Decls {
			decl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue // *ast.FuncDecl
			}

			switch decl.Tok {
			case token.CONST, token.VAR:
				for _, spec := range decl.Specs {
					spec := spec.(*ast.ValueSpec)

					checker.CheckValueSpec(spec, decl)
				}

			case token.TYPE:
				for _, spec := range decl.Specs {
					spec := spec.(*ast.TypeSpec)

					checker.CheckTypeSpec(spec, decl)
				}

			default:
				continue // token.IMPORT
			}
		}
	}

	return checker.Violations()
}
