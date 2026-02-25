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

package gendoc

import (
	"errors"
	"flag"
	"strings"
	"testing"
)

func TestSplitDoc(t *testing.T) {
	t.Parallel()

	const doc = `reports examples

The example analyzer reports examples.

	example ./...

# Examples

More examples.
`

	got, err := splitDoc(doc)
	if err != nil {
		t.Fatal(err)
	}

	if want := (docSections{
		Summary: "reports examples",
		Body:    "The example analyzer reports examples.\n\n\texample ./...",
		Rest:    "# Examples\n\nMore examples.",
	}); got != want {
		t.Errorf("splitDoc() = %#v, want %#v", got, want)
	}
}

func TestSplitDocSummaryOnly(t *testing.T) {
	t.Parallel()

	got, err := splitDoc("reports examples\n")
	if err != nil {
		t.Fatal(err)
	}

	if want := (docSections{Summary: "reports examples"}); got != want {
		t.Errorf("splitDoc() = %#v, want %#v", got, want)
	}
}

func TestSplitDocErrors(t *testing.T) {
	t.Parallel()

	tests := [...]struct{ name, doc string }{
		{"empty", ""},
		{"multi-line summary", "reports\nexamples\n\nThe example analyzer."},
		{"heading", "# Examples\n\nMore examples."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := splitDoc(tt.doc); !errors.Is(err, errSummary) {
				t.Errorf("splitDoc() error = %v, want %v", err, errSummary)
			}
		})
	}
}

func TestRenderFlags(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Bool("verbose", true, "print more")
	fs.String("config", "", "read the configuration from `file`")

	got, err := renderFlags(fs)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range [...]string{
		"\t-c int\n\t\tdisplay offending line with this many lines of context (default -1)\n",
		"\t-config file\n\t\tread the configuration from file\n",
		"\t-verbose\n\t\tprint more (default true)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("renderFlags() = %q, want it to contain %q", got, want)
		}
	}
}

func TestRenderFlagsDuplicate(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Bool("fix", false, "fix it")

	if _, err := renderFlags(fs); !errors.Is(err, errDuplicateFlag) {
		t.Errorf("renderFlags() error = %v, want %v", err, errDuplicateFlag)
	}
}

func TestRenderCommandCommentEnd(t *testing.T) {
	t.Parallel()

	if _, err := renderCommand("example", docSections{Summary: "reports */ examples"}, ""); !errors.Is(err, errCommentEnd) {
		t.Errorf("renderCommand() error = %v, want %v", err, errCommentEnd)
	}
}
