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

package errorwrapper

import (
	"context"
	"flag"
	"fmt"
	"regexp"

	"fillmore-labs.com/errortype/internal/detectwrappers/result"
	"fillmore-labs.com/errortype/internal/detectwrappers/wrappers"
	"fillmore-labs.com/errortype/internal/overrides"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// Config define configuration settings for type analysis, including heuristic passes, usage overrides, and tracing.
type Config struct {
	WrapperOverrides WrapperOverrides
	Trace            *regexp.Regexp
}

// WrapperOverrides maps a package path to a map of function names to their corresponding error wrapper metadata.
type WrapperOverrides map[string][]wrappers.ExplicitWrapper

// DefaultConfig returns a [Config] struct initialized with default values.
func DefaultConfig() Config {
	return Config{ // Default options
		Trace: nil,
	}
}

// RegisterFlags binds the [Config] values to command line flag values.
func (o *Config) RegisterFlags(*flag.FlagSet) {}

// ReadOverrides reads wrapper function overrides from the specified file.
// If fileName is empty, no action is taken.
func (o *Config) ReadOverrides(fileName string) error {
	if fileName == "" {
		return nil
	}

	ctx := context.Background()

	usageOverrides, err := overrides.ReadWrappersFile(ctx, fileName)
	if err != nil {
		return fmt.Errorf("can't read overrides file %s: %w", fileName, err)
	}

	o.AddWrappers(usageOverrides.Wrappers)

	return nil
}

// AddWrappers merges the given overrides into the WrapperOverrides map.
func (o *Config) AddWrappers(or map[result.WrapperType][]typeutil.FuncName) {
	if len(or) == 0 {
		return
	}

	if o.WrapperOverrides == nil {
		o.WrapperOverrides = make(WrapperOverrides)
	}

	for wrapperType, names := range or {
		for _, name := range names {
			pkg := name.Path
			o.WrapperOverrides[pkg] = append(o.WrapperOverrides[pkg], wrappers.ExplicitWrapper{LocalFuncName: name.LocalFuncName, Type: wrapperType})
		}
	}
}

// SetTrace sets the Trace field to a compiled regular expression based on the provided regex string.
func (o *Config) SetTrace(regex string) error {
	if regex == "" {
		o.Trace = nil

		return nil
	}

	re, err := regexp.Compile(regex)
	if err != nil {
		return err
	}

	o.Trace = re

	return nil
}
