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

package detectwrappers

import (
	"reflect"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/detectwrappers/result"
	"fillmore-labs.com/errortype/internal/settings/errorwrapper"
)

// NewAnalyzer creates a detectwrappers *[analysis.NewAnalyzer].
// It detects wrapper function to provide
// this information to other analyzers in the toolchain.
func NewAnalyzer(o *errorwrapper.Config) *analysis.Analyzer {
	a := &analysis.Analyzer{
		Name:             "detectwrappers",
		Doc:              "Determines how error types are used (pointer vs. value) for use by other analyzers.",
		URL:              "https://pkg.go.dev/fillmore-labs.com/errortype/detect",
		Run:              func(p *analysis.Pass) (any, error) { return Run(p, o) },
		RunDespiteErrors: true,
		ResultType:       reflect.TypeFor[result.Result](),
		FactTypes:        []analysis.Fact{(*result.ErrorFunc)(nil)},
	}

	o.RegisterFlags(&a.Flags)

	return a
}
