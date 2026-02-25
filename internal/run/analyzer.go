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

package run

import (
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"fillmore-labs.com/errortype/internal/settings"
)

// Documentation constants.
const (
	name = "errortype"
	url  = "https://pkg.go.dev/fillmore-labs.com/errortype"
)

// NewAnalyzer creates an *[analysis.Analyzer] instance for identifying error handling issues,
// documented by doc.
func NewAnalyzer(o *settings.Config, doc string) *analysis.Analyzer {
	a := &analysis.Analyzer{
		Name:     name,
		Doc:      strings.TrimSpace(doc),
		URL:      url,
		Run:      func(p *analysis.Pass) (any, error) { return Run(p, o) },
		Requires: []*analysis.Analyzer{inspect.Analyzer, o.TypeAnalyzer, o.WrapperAnalyzer},
	}

	if o.TypeAnalyzer == nil || o.WrapperAnalyzer == nil {
		panic("nil")
	}

	o.RegisterFlags(&a.Flags)

	return a
}
