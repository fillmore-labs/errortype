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

package overrides

import (
	"fillmore-labs.com/errortype/detect/result"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// Overrides associates an error type with a list of fully qualified type names,
// and wrapper functions, methods, or variables to their target names.
type Overrides struct {
	Types    map[result.ErrorType][]typeutil.TypeName
	Wrappers map[result.WrapperType][]typeutil.FuncName
}

// errorFileType represents the configuration for error type overrides in files.
//
// It categorizes type names into five groups.
type errorFileType struct {
	Wrappers     map[result.WrapperType][]typeutil.FuncName `yaml:"wrappers,omitempty"`
	Pointer      []typeutil.TypeName                        `yaml:"pointer,omitempty"`
	Value        []typeutil.TypeName                        `yaml:"value,omitempty"`
	Suppress     []typeutil.TypeName                        `yaml:"suppress,omitempty"`
	Inconsistent []typeutil.TypeName                        `yaml:"inconsistent,omitempty"`
}
