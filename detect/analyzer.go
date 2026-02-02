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

package detect

import (
	"reflect"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/facts"
	"fillmore-labs.com/errortype/internal/detect"
)

// New creates a new instance of the detecttypes analyzer.
// It detects how error types are used (as pointers or values) to provide
// this information to other analyzers in the toolchain.
func New(opts ...Option) *analysis.Analyzer {
	o := detect.DefaultOptions()
	Options(opts).apply(o)

	a := &analysis.Analyzer{
		Name:             "detecttypes",
		Doc:              "Determines how error types are used (pointer vs. value) for use by other analyzers.",
		URL:              "https://pkg.go.dev/fillmore-labs.com/errortype/detect",
		Run:              o.Run,
		RunDespiteErrors: true,
		FactTypes:        []analysis.Fact{(*facts.ErrorFact)(nil)},
		ResultType:       reflect.TypeFor[facts.Result](),
	}

	registerFlags(o, &a.Flags)

	return a
}

// Analyzer is a pre-configured *[analysis.Analyzer] for detecting error types in Go programs.
var Analyzer = New()
