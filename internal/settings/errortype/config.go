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

package errortype

import (
	"context"
	"flag"
	"fmt"
	"regexp"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/detecttypes/result"
	"fillmore-labs.com/errortype/internal/overrides"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// Config define configuration settings for type analysis, including heuristic passes, usage overrides, and tracing.
type Config struct {
	UsageOverrides  UsageOverrides
	WrapperAnalyzer *analysis.Analyzer
	Trace           *regexp.Regexp
	Heuristics      Heuristic
}

// UsageOverrides maps a fully qualified type name to its corresponding error usage type.
type UsageOverrides map[typeutil.TypeName]result.ErrorType

// DefaultConfig returns a [Config] struct initialized with default values.
func DefaultConfig() Config {
	return Config{ // Default options
		UsageOverrides: nil,
		Heuristics:     HeuristicAll,
		Trace:          nil,
	}
}

// RegisterFlags binds the [Config] values to command line flag values.
func (o *Config) RegisterFlags(fs *flag.FlagSet) {
	fs.TextVar(&o.Heuristics, "heuristics", o.Heuristics, "list of `heuristics` used; values: \"var\", \"usage\", \"receivers\", \"off\" to disable")
}

// ReadOverrides reads error type usage overrides from the specified file.
// If fileName is empty, no action is taken.
func (o *Config) ReadOverrides(fileName string) error {
	if fileName == "" {
		return nil
	}

	ctx := context.Background()

	usageOverrides, err := overrides.ReadTypesFile(ctx, fileName)
	if err != nil {
		return fmt.Errorf("can't read overrides file %s: %w", fileName, err)
	}

	o.AddOverrides(usageOverrides.Types)

	return nil
}

// AddOverrides merges the given overrides into the UsageOverrides map.
func (o *Config) AddOverrides(or map[result.ErrorType][]typeutil.TypeName) {
	if len(or) == 0 {
		return
	}

	if o.UsageOverrides == nil {
		o.UsageOverrides = make(UsageOverrides)
	}

	for typ, names := range or {
		for _, name := range names {
			o.UsageOverrides[name] = typ
		}
	}
}
