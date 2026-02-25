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
	"flag"
	"reflect"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/detect/result"
)

// Analyzer creates a detecttypes *[analysis.Analyzer] to determine how error types are used (pointer vs. value).
func (o *Options) Analyzer() *analysis.Analyzer {
	a := &analysis.Analyzer{
		Name:             "detecttypes",
		Doc:              "Determines how error types are used (pointer vs. value) for use by other analyzers.",
		URL:              "https://pkg.go.dev/fillmore-labs.com/errortype/detect",
		Run:              o.Run,
		RunDespiteErrors: true,
		ResultType:       reflect.TypeFor[result.Result](),
		FactTypes:        []analysis.Fact{(*result.ErrorType)(nil), (*result.ErrorFunc)(nil)},
	}

	o.registerFlags(&a.Flags)

	return a
}

// registerFlags binds the [Options] values to command line flag values.
func (o *Options) registerFlags(fs *flag.FlagSet) {
	fs.Func("overrides", "read error type overrides from this `file`", o.ReadOverrides)
	fs.TextVar(&o.Heuristics, "heuristics", o.Heuristics, "list of `heuristics` used; values: \"var\", \"usage\", \"receivers\", \"off\" to disable")
	fs.Func("tracetypes", "information of error type detection in packages matching this `regex`", o.SetTrace)
}
