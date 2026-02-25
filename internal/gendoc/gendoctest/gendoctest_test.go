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

package gendoctest

import (
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/gendoc"
)

func TestCheckUpToDate(t *testing.T) {
	a := &analysis.Analyzer{Name: "example", Doc: "reports examples"}

	generated, err := gendoc.Generate(a.Name, a.Doc, &a.Flags)
	if err != nil {
		t.Fatal(err)
	}

	stale := strings.Replace(string(generated), "reports examples", "reports something", 1)

	tests := [...]struct {
		name, content string
		wantErr       error
		wantMsg       string
	}{
		{"up to date", string(generated), nil, ""},
		{"stale", stale, ErrOutOfDate, "something"},
		{"not generated", "package analyzer\n", gendoc.ErrNotGenerated, gendoc.CommandFile},
		{"missing", "", fs.ErrNotExist, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(t.TempDir())

			if err := os.Mkdir("analyzer", 0o700); err != nil {
				t.Fatal(err)
			}

			if tt.content != "" {
				if err := os.WriteFile("doc.go", []byte(tt.content), 0o600); err != nil {
					t.Fatal(err)
				}
			}

			t.Chdir("analyzer")

			err := CheckUpToDate(a)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CheckUpToDate() = %v, want %v", err, tt.wantErr)
			}

			if err != nil && !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("CheckUpToDate() = %v, want it to contain %q", err, tt.wantMsg)
			}
		})
	}
}
