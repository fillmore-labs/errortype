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

import "regexp"

// Options define configuration settings for type analysis, including heuristic passes, usage overrides, and tracing.
type Options struct {
	// UsageOverrides stores the usage configuration for error types, read from a file.
	UsageOverrides

	// WrapperOverrides stores the wrapper function configuration, read from a file.
	WrapperOverrides

	// Heuristics controls heuristic passes
	Heuristics

	// Trace controls result output
	Trace *regexp.Regexp

	// InitializationError is set when [Options] intialization fails
	InitializationError error
}

// DefaultOptions returns a [Options] struct initialized with default values.
func DefaultOptions() *Options {
	return &Options{ // Default options
		UsageOverrides:      nil,
		Heuristics:          HeuristicAll,
		Trace:               nil,
		InitializationError: nil,
	}
}
