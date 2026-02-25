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
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"fillmore-labs.com/errortype/internal/settings"
)

// Documentation constants.
const (
	name = "errortype"
	doc  = `errortype is a static analysis tool that helps prevent subtle bugs in error handling.

It performs three checks:

1. Inconsistent Error Type Usage: Ensures error types are used consistently
   as either pointers or values in returns, type assertions, and errors.As calls.

2. Pointless Comparisons: Detects comparisons against newly allocated addresses
   (like errors.Is(err, &url.Error{}) or ptr == &MyStruct{}), which are almost always incorrect.

3. Error Naming Conventions (opt-in): Checks that sentinel error variables use the Err prefix (e.g.,
   ErrNotFound) and structured error types use the Error suffix (e.g., ParseError).

For inconsistent error type usage, it automatically determines the correct usage
for most error types but may require a configuration file for ambiguous cases.`

	url = "https://pkg.go.dev/fillmore-labs.com/errortype"
)

// NewAnalyzer creates an *[analysis.NewAnalyzer] instance for identifying error handling issues.
func NewAnalyzer(o *settings.Config) *analysis.Analyzer {
	a := &analysis.Analyzer{
		Name:     name,
		Doc:      doc,
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
