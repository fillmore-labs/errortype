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

package detecttypes

import (
	"reflect"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/detecttypes/result"
	"fillmore-labs.com/errortype/internal/detectwrappers"
	"fillmore-labs.com/errortype/internal/settings/errortype"
	"fillmore-labs.com/errortype/internal/settings/errorwrapper"
)

// NewAnalyzer creates a detecttypes *[analysis.NewAnalyzer] to determine how error types are used (pointer vs. value).
func NewAnalyzer(o *errortype.Config) *analysis.Analyzer {
	if o.WrapperAnalyzer == nil {
		ow := errorwrapper.DefaultConfig()

		o.WrapperAnalyzer = detectwrappers.NewAnalyzer(&ow)
	}

	a := &analysis.Analyzer{
		Name:             "detecttypes",
		Doc:              "Determines how error types are used (pointer vs. value) for use by other analyzers.",
		URL:              "https://pkg.go.dev/fillmore-labs.com/errortype/detect",
		Run:              func(p *analysis.Pass) (any, error) { return Run(p, o) },
		RunDespiteErrors: true,
		Requires:         []*analysis.Analyzer{o.WrapperAnalyzer},
		ResultType:       reflect.TypeFor[result.Result](),
		FactTypes:        []analysis.Fact{(*result.ErrorType)(nil)},
	}

	o.RegisterFlags(&a.Flags)

	return a
}
