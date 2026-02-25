// Copyright 2026 Oliver Eikemeier. All Rights Reserved.
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

package settings

import (
	"flag"
	"regexp"
	"sync"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/settings/errortype"
	"fillmore-labs.com/errortype/internal/settings/errorwrapper"
)

// Config provides configurations for analysis passes, including type detection and AST-related behavior customization.
type Config struct {
	WrapperConfig errorwrapper.Config
	TypeConfig    errortype.Config

	TypeAnalyzer *analysis.Analyzer

	WrapperAnalyzer *analysis.Analyzer

	// Suggest appends suggestions to a file
	Suggest      string
	SuggestWrite sync.Mutex
	Options      Options
}

// RegisterFlags binds the [analyze.RunOptions] values to command line flag values.
func (o *Config) RegisterFlags(fs *flag.FlagSet) {
	for _, flag := range [...]struct {
		usage string
		mask  Options
	}{
		{usage: `analyze generated files`, mask: OptionGenerated},
		{usage: `extended checks for not comparable error types`, mask: OptionNotComparable},
		{usage: `check error sentinel and type names`, mask: OptionNaming},
		{usage: `check for pre-Go1.13 error query helpers`, mask: OptionLegacy},
		{usage: `check for confusing uses of errors.As`, mask: OptionStyleCheck},
		{usage: `report unchecked calls on errors.As-like functions`, mask: OptionCheckUnused},
		{usage: `suppress compare diagnostic on errors.Is if the compared type has an "Is(error) bool" method`, mask: OptionCheckIs},
		{usage: `diagnose all "Unwrap" functions in "Is" methods, not only those on target`, mask: OptionDeepIsCheck},
		{usage: `report unchecked type asserts on errors`, mask: OptionUncheckedAssert},
		{usage: `restrict variable analysis to variables with an "err" prefix`, mask: OptionPrefixFilter},
	} {
		usage, mask := flag.usage, flag.mask
		fs.Var(o.Options.BoolFlag(mask), mask.String(), usage)
	}

	fs.BoolFunc("recommended", `set recommended options`, o.Options.Recommended)

	fs.StringVar(&o.Suggest, "suggest", o.Suggest, "append suggestions to this `file`, \"-\" for standard output")

	fs.Func("overrides", "read error type overrides from this `file`", o.ReadOverrides)
	fs.Func("tracetypes", "information of error type detection in packages matching this `regex`", o.SetTrace)

	o.TypeConfig.RegisterFlags(fs)
	o.WrapperConfig.RegisterFlags(fs)
}

// ReadOverrides reads error type and wrapper usage overrides from the specified file.
// If fileName is empty, no action is taken.
func (o *Config) ReadOverrides(fileName string) error {
	if err := o.TypeConfig.ReadOverrides(fileName); err != nil {
		return err
	}

	return o.WrapperConfig.ReadOverrides(fileName)
}

// SetTrace sets the Trace field to a compiled regular expression based on the provided regex string.
func (o *Config) SetTrace(regex string) error {
	if regex == "" {
		o.WrapperConfig.Trace = nil
		o.TypeConfig.Trace = nil

		return nil
	}

	re, err := regexp.Compile(regex)
	if err != nil {
		return err
	}

	o.WrapperConfig.Trace = re
	o.TypeConfig.Trace = re

	return nil
}

// DefaultConfig returns a [Config] struct initialized with default values.
func DefaultConfig() *Config {
	return &Config{ // Default options
		Suggest:       "",
		Options:       DefaultOptions,
		TypeConfig:    errortype.DefaultConfig(),
		WrapperConfig: errorwrapper.DefaultConfig(),
	}
}
