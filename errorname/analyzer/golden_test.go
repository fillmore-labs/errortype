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

package analyzer_test

import (
	"go/ast"
	"go/build"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/txtar"
)

// TestGoldenTypeChecks verifies that the suggested fixes produce a valid
// package: every testdata source file is type-checked together with the
// others, substituting the .golden variant — the post-fix state — when one
// exists.
func TestGoldenTypeChecks(t *testing.T) {
	t.Parallel()

	dir := analysistest.TestData()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()

	var files [][]namedAST

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}

		match, err := build.Default.MatchFile(dir, name)
		if err != nil {
			t.Fatalf("match file %s: %v", name, err)
		}

		if !match {
			continue
		}

		path := filepath.Join(dir, name)

		alternatives := parseFiles(t, fset, path)

		files = append(files, alternatives)
	}

	checkCombinations(t, fset, files, 0, nil)
}

func parseFiles(tb testing.TB, fset *token.FileSet, path string) []namedAST {
	tb.Helper()

	var alternatives []namedAST
	golden := path + ".golden"

	if !fileExists(golden) {
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			tb.Fatalf("parse %s: %v", path, err)
		}

		alternatives = append(alternatives, namedAST{name: path, file: file})

		return alternatives
	}

	ar, err := txtar.ParseFile(golden)
	if err != nil {
		tb.Fatalf("parse txtar %s: %v", golden, err)
	}

	if len(ar.Files) > 0 {
		for _, section := range ar.Files {
			sectionPath := golden + "[" + section.Name + "]"

			file, err := parser.ParseFile(fset, sectionPath, section.Data, parser.SkipObjectResolution)
			if err != nil {
				tb.Fatalf("parse %s section %s: %v", golden, section.Name, err)
			}

			alternatives = append(alternatives, namedAST{file: file, name: sectionPath})
		}

		return alternatives
	}

	file, err := parser.ParseFile(fset, golden, nil, parser.SkipObjectResolution)
	if err != nil {
		tb.Fatalf("parse %s: %v", golden, err)
	}

	alternatives = append(alternatives, namedAST{file: file, name: golden})

	return alternatives
}

type namedAST struct {
	file *ast.File
	name string
}

// var checkCombinations func(int, []namedAST).
func checkCombinations(tb testing.TB, fset *token.FileSet, files [][]namedAST, index int, current []namedAST) {
	tb.Helper()

	if index == len(files) {
		var astFiles []*ast.File
		for _, na := range current {
			astFiles = append(astFiles, na.file)
		}
		conf := types.Config{
			Importer: importer.Default(),
			Error:    func(err error) { tb.Error(err) },
		}
		_, _ = conf.Check("testdata", fset, astFiles, nil)

		return
	}

	for _, na := range files[index] {
		checkCombinations(tb, fset, files, index+1, append(current, na))
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
