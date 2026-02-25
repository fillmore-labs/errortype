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
	"context"
	"fmt"
	"regexp"

	"fillmore-labs.com/errortype/detect/result"
	"fillmore-labs.com/errortype/internal/overrides"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// ReadOverrides reads error type usage overrides from the specified file.
// If fileName is empty, no action is taken.
func (o *Options) ReadOverrides(fileName string) error {
	if fileName == "" {
		return nil
	}

	ctx := context.Background()

	usageOverrides, err := overrides.ReadFile(ctx, fileName)
	if err != nil {
		return fmt.Errorf("can't read overrides file %s: %w", fileName, err)
	}

	o.AddOverrides(usageOverrides.Types)
	o.AddWrappers(usageOverrides.Wrappers)

	return nil
}

// AddOverrides merges the given overrides into the UsageOverrides map.
func (o *Options) AddOverrides(or map[result.ErrorType][]typeutil.TypeName) {
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

// AddWrappers merges the given overrides into the WrapperOverrides map.
func (o *Options) AddWrappers(or map[result.WrapperType][]typeutil.FuncName) {
	if len(or) == 0 {
		return
	}

	if o.WrapperOverrides == nil {
		o.WrapperOverrides = make(WrapperOverrides)
	}

	for wrapperType, names := range or {
		for _, name := range names {
			o.WrapperOverrides[name] = wrapperType
		}
	}
}

// SetHeuristics parses and sets the heuristic passes from a comma-separated list.
// Valid values are: "usage", "receivers", and "off".
// "off" disables all heuristics and cannot be combined with other values.
func (o *Options) SetHeuristics(list string) error {
	// Only update if the user provided some values.
	if list == "" {
		return nil
	}

	heuristics, err := HeuristicsFromString(list)
	if err != nil {
		return err
	}

	o.Heuristics = heuristics

	return nil
}

// SetTrace sets the Trace field to a compiled regular expression based on the provided regex string.
func (o *Options) SetTrace(regex string) error {
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
