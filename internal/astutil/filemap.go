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

package astutil

import (
	"go/ast"
	"go/token"
	"os"
	"slices"
)

// FileMap indexes the files of an analysis pass by their [token.File],
// caching per-file metadata for the position-based queries.
type FileMap struct {
	fset      *token.FileSet
	ast       map[*token.File]*ast.File
	generated map[*ast.File]struct{}
}

// NewFileMap creates a [FileMap] that associates files in a [token.FileSet] with their
// corresponding [ast.File] objects.
func NewFileMap(fset *token.FileSet, files []*ast.File) FileMap {
	var generated map[*ast.File]struct{}

	astFiles := make(map[*token.File]*ast.File, len(files))
	for _, astFile := range files {
		if ast.IsGenerated(astFile) {
			if generated == nil {
				generated = make(map[*ast.File]struct{})
			}
			generated[astFile] = struct{}{}
		}

		if tokenFile := fset.File(astFile.FileStart); tokenFile != nil {
			astFiles[tokenFile] = astFile
		}
	}

	return FileMap{
		fset:      fset,
		ast:       astFiles,
		generated: generated,
	}
}

// File is the [ast.File] where the [token.Pos] is in.
func (f FileMap) File(pos token.Pos) *ast.File {
	tokenFile := f.fset.File(pos)
	return f.ast[tokenFile]
}

// IsGenerated reports whether file is generated.
func (f FileMap) IsGenerated(file *ast.File) bool {
	_, ok := f.generated[file]
	return ok
}

// InGenerated reports whether pos is in a generated file.
func (f FileMap) InGenerated(pos token.Pos) bool {
	if len(f.generated) == 0 {
		return false
	}

	return f.IsGenerated(f.File(pos))
}

// HasGenerated checks if any of the specified positions is in a generated file.
func (f FileMap) HasGenerated(positions []token.Pos) bool {
	if len(f.generated) == 0 {
		return false
	}

	return slices.ContainsFunc(positions, f.InGenerated)
}

// Print is useful for debugging.
func (f FileMap) Print(x any) error {
	return ast.Fprint(os.Stdout, f.fset, x, ast.NotNilFilter)
}
