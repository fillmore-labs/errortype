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
	"go/token"
	"go/types"
	"strings"
)

// PackageLevel reports whether obj is declared at package scope.
func PackageLevel(obj types.Object) bool {
	return obj.Parent() == obj.Pkg().Scope()
}

// PkgPath returns the package path of the given [analysis.Pass], appending “(test)” if it belongs to a test package.
func PkgPath(fset *token.FileSet, pkg *types.Package) string {
	pkgPath := pkg.Path()
	if isTest(fset, pkg) {
		pkgPath += " (test)"
	}

	return pkgPath
}

// isTest determines if the given package is a test package, either external or containing test files.
func isTest(fset *token.FileSet, pkg *types.Package) bool {
	if strings.HasSuffix(pkg.Path(), "_test") {
		return true // This is an external test package
	}

	// Check if any files in the package are test files
	for file := range fset.Iterate {
		if strings.HasSuffix(file.Name(), "_test.go") {
			return true
		}
	}

	return false
}
