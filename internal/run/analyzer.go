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
	"flag"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"

	"fillmore-labs.com/errortype/internal/analyze"
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

// Analyzer creates an *[analysis.Analyzer] instance for identifying error handling issues.
func (o *Options) Analyzer() *analysis.Analyzer {
	a := &analysis.Analyzer{
		Name:     name,
		Doc:      doc,
		URL:      url,
		Run:      o.Run,
		Requires: []*analysis.Analyzer{inspect.Analyzer, o.DetectTypes},
	}

	o.registerFlags(&a.Flags)

	return a
}

// registerFlags binds the [analyze.RunOptions] values to command line flag values.
func (o *Options) registerFlags(fs *flag.FlagSet) {
	for _, flag := range [...]struct {
		usage string
		mask  analyze.Options
	}{
		{usage: `analyze generated files`, mask: analyze.OptionGenerated},
		{usage: `extended checks for not comparable error types`, mask: analyze.OptionNotComparable},
		{usage: `check error sentinel and type names`, mask: analyze.OptionNaming},
		{usage: `check for pre-Go1.13 error query helpers`, mask: analyze.OptionLegacy},
		{usage: `check for confusing uses of errors.As`, mask: analyze.OptionStyleCheck},
		{usage: `report unchecked calls on errors.As-like functions`, mask: analyze.OptionCheckUnused},
		{usage: `suppress compare diagnostic on errors.Is if the compared type has an "Is(error) bool" method`, mask: analyze.OptionCheckIs},
		{usage: `diagnose all "Unwrap" functions in "Is" methods, not only those on target`, mask: analyze.OptionDeepIsCheck},
		{usage: `report unchecked type asserts on errors`, mask: analyze.OptionUncheckedAssert},
		{usage: `restrict variable analysis to variables with an "err" prefix`, mask: analyze.OptionPrefixFilter},
	} {
		usage, mask := flag.usage, flag.mask
		fs.Var(o.Options.BoolFlag(mask), mask.String(), usage)
	}

	fs.BoolFunc("recommended", `set recommended options`, o.Options.Recommended)

	fs.StringVar(&o.Suggest, "suggest", o.Suggest, "append suggestions to this `file`, \"-\" for standard output")

	copyFlags(&o.DetectTypes.Flags, fs)
}

// copyFlags copies all flags from one FlagSet to another.
// It panics if a flag with the same name already exists in the destination FlagSet,
// as this indicates a programming error in the analyzer setup.
func copyFlags(from, to *flag.FlagSet) {
	from.VisitAll(func(f *flag.Flag) {
		if to.Lookup(f.Name) != nil { // programming error
			panic("duplicate flag " + f.Name)
		}

		to.Var(f.Value, f.Name, f.Usage)
	})
}
