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
	"sync"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/analyze"
)

// Options provide configurations for analysis passes, including type detection and AST-related behavior customization.
type Options struct {
	analyze.Options

	DetectTypes *analysis.Analyzer

	// Suggest appends suggestions to a file
	Suggest string

	suggestwrite sync.Mutex
}

// DefaultRunOptions returns a [Options] struct initialized with default values.
func DefaultRunOptions() *Options {
	return &Options{ // Default options
		Options:     analyze.DefaultOptions,
		DetectTypes: nil,
		Suggest:     "",
	}
}
