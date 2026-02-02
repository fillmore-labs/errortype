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

package analyzer

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"fillmore-labs.com/errortype/detect"
	"fillmore-labs.com/errortype/internal/run"
)

// Documentation constants.
const (
	Name = "errortype"
	Doc  = `errortype is a Go static analysis tool that helps prevent subtle bugs in error handling.

It performs two main checks:

1. Inconsistent Error Type Usage: It analyzes function returns, type assertions,
   and calls to functions like errors.As to ensure that custom error types
   are used consistently as either pointers or values.

2. Pointless Pointer Comparisons: It detects comparisons of pointers against
   the address of a newly created value (e.g., 'ptr == &MyStruct{}'), which
   are almost always incorrect.

For inconsistent error type usage, it automatically determines the correct usage
for most error types but may require a configuration file for ambiguous cases.`

	URL = "https://pkg.go.dev/fillmore-labs.com/errortype/analyzer"
)

// New creates a new instance of the errortype analyzer.
// It allows for programmatic configuration using [Options], which is useful
// for integrating the analyzer into other tools. For command-line use, the
// pre-configured [Analyzer] variable is typically sufficient.
func New(opts ...Option) *analysis.Analyzer {
	r := run.DefaultRunOptions()
	Options(opts).apply(r)

	if r.DetectTypes == nil {
		r.DetectTypes = detect.New()
	}

	a := &analysis.Analyzer{
		Name:     Name,
		Doc:      Doc,
		URL:      URL,
		Run:      r.Run,
		Requires: []*analysis.Analyzer{inspect.Analyzer, r.DetectTypes},
	}

	registerFlags(r, &a.Flags)

	return a
}

// Analyzer is a pre-configured *[analysis.Analyzer] for detecting and enforcing consistent error type usage in Go programs.
var Analyzer = New(WithDetectTypes(detect.Analyzer))
