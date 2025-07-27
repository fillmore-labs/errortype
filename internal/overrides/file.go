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
	"fillmore-labs.com/errortype/internal/errortypes"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// errorfileType represents the configuration for error type overrides in files.
//
// It categorizes type names into four groups.
type errorfileType struct {
	// Types that should be treated as pointer errors.
	Pointer []typeutil.TypeName `yaml:"pointer,omitempty"`
	// Types that should be treated as value errors.
	Value []typeutil.TypeName `yaml:"value,omitempty"`
	// Types for which error type checks should be suppressed - never written.
	Suppress []typeutil.TypeName `yaml:"suppress,omitempty"`
	//  Types that have inconsistent error type usage - ignored on read.
	Inconsistent []typeutil.TypeName `yaml:"inconsistent,omitempty"`
}

// Override represents a mapping between a Go type and its associated error type.
// It combines a TypeName with an ErrorType for error handling customization.
type Override struct {
	typeutil.TypeName
	errortypes.ErrorType
}
