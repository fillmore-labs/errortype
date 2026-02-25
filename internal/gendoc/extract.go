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
	"strings"
)

// docSections are the parts of the analyzer documentation.
type docSections struct {
	Summary string // the first paragraph, a single line
	Body    string // the description up to the first heading
	Rest    string // the first heading and everything after it
}

var errSummary = errors.New("documentation does not start with a single-line summary")

// splitDoc splits the analyzer documentation into its sections.
func splitDoc(doc string) (docSections, error) {
	summary, text, _ := strings.Cut(strings.TrimSpace(doc), "\n\n")
	if summary == "" || strings.Contains(summary, "\n") || strings.HasPrefix(summary, "# ") {
		return docSections{}, errSummary
	}

	text = "\n" + text

	body, rest, found := strings.Cut(text, "\n# ")
	if found {
		rest = "# " + rest
	}

	return docSections{
		Summary: summary,
		Body:    strings.TrimSpace(body),
		Rest:    strings.TrimSpace(rest),
	}, nil
}
