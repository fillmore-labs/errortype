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

	"fillmore-labs.com/errortype/internal/errortypes"
)

// Analyzer creates an instance of the detecttypes analyzer.
// It detects how error types are used (as pointers or values) to provide
// this information to other analyzers in the toolchain.
func (o *Options) Analyzer() *analysis.Analyzer {
	a := &analysis.Analyzer{
		Name:             "detecttypes",
		Doc:              "Determines how error types are used (pointer vs. value) for use by other analyzers.",
		URL:              "https://pkg.go.dev/fillmore-labs.com/errortype/detect",
		Run:              o.run,
		RunDespiteErrors: true,
		FactTypes:        []analysis.Fact{(*errortypes.ErrorType)(nil)},
		ResultType:       reflect.TypeFor[errortypes.Result](),
	}

	a.Flags.Func("overrides", "read error type overrides from this file", o.ReadOverrides)
	a.Flags.Func("heuristics", `list of heuristics used (default: "var,usage,receivers", "off" to disable)`, o.SetHeuristics)
	a.Flags.Func("tracetypes", "information of error type detection in packages matching this regex", o.SetTrace)

	return a
}
