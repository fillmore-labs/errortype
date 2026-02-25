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

package typeutil

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"
)

// PackageLevel reports whether obj is declared at package scope.
func PackageLevel(obj types.Object) bool {
	return obj.Parent() == obj.Pkg().Scope()
}

// PkgPath returns the path of pkg, with “ (test)” appended when it is a test package.
// files are the package's syntax files, as in [analysis.Pass.Files].
func PkgPath(fset *token.FileSet, pkg *types.Package, files []*ast.File) string {
	pkgPath := pkg.Path()
	if isTest(fset, pkg, files) {
		pkgPath += " (test)"
	}

	return pkgPath
}

// isTest reports whether pkg is an external test package or files contain test files.
//
// Only files is checked, not all of fset: drivers like singlechecker share one
// [token.FileSet] across all loaded packages.
func isTest(fset *token.FileSet, pkg *types.Package, files []*ast.File) bool {
	if strings.HasSuffix(pkg.Path(), "_test") {
		return true // external test package
	}

	for _, f := range files {
		if strings.HasSuffix(fset.File(f.FileStart).Name(), "_test.go") {
			return true
		}
	}

	return false
}
