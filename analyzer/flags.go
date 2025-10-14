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
	"flag"
	"strconv"

	"fillmore-labs.com/errortype/internal/analyze"
)

// registerFlags binds the [analyze.RunOptions] values to command line flag values.
// A nil flag set value defaults to the program's command line.
func registerFlags(o *analyze.RunOptions, fs *flag.FlagSet) {
	if fs == nil {
		fs = flag.CommandLine
	}

	flags := [...]struct {
		name, usage string
		flag        analyze.Options
	}{
		{"style-check", `check for confusing uses of errors.As`, analyze.OptionStyleCheck},
		{"check-is", `suppress compare diagnostic on errors.Is if the compared type has an "Is(error) bool" method`, analyze.OptionCheckIs},
		{"deep-is-check", `diagnose all "Unwrap" functions in "Is" methods, not only those on target`, analyze.OptionDeepIsCheck},
		{"unchecked-assert", `report unchecked type asserts on errors`, analyze.OptionUncheckedAssert},
		{"check-unused", `report unchecked calls on errors.As-like functions`, analyze.OptionCheckUnused},
	}

	for _, f := range flags {
		fs.Var(optionValue{o: &o.Options, flag: f.flag}, f.name, f.usage)
	}

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

// optionValue implements [flag.Value] to bind an [Options] configuration to a specific flag for command-line parsing.
type optionValue struct {
	o    *analyze.Options
	flag analyze.Options
}

// String implements [flag.Value].
func (v optionValue) String() string {
	return strconv.FormatBool(v.o != nil && *v.o&v.flag != 0)
}

// Set implements [flag.Value].
func (v optionValue) Set(s string) error {
	b, err := strconv.ParseBool(s)
	if v.o != nil && err == nil {
		v.o.SetOption(v.flag, b)
	}

	return err
}

// IsBoolFlag makes `-name` equivalent to `-name=true`.
func (v optionValue) IsBoolFlag() bool { return true }
