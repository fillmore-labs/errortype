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

package analyze

import (
	"sync"

	"golang.org/x/tools/go/analysis"
)

// AstOptions represents configuration flags to control the behavior of style and correctness checks for errors.
type AstOptions struct {
	StyleCheck bool // StyleCheck controls the target style check in `errors.As` calls

	CheckIs bool // CheckIs controls whether to check for `Is(error) bool` methods

	DeepIsCheck bool // DeepIsCheck flags all unwrap methods in `Is` method checks, not only those on target

	UncheckedAssert bool // UncheckedAssert flags all uncheckd asserts on errors

	CheckUnused bool // CheckUnused flags unchecked results of `errors.As` calls
}

// Options provide configurations for analysis passes, including type detection and AST-related behavior customization.
type Options struct {
	DetectTypes *analysis.Analyzer

	AstOptions

	Suggest string // Suggest appends suggestions to a file

	suggestwrite sync.Mutex
}

// DefaultOptions returns a [Options] struct initialized with default values.
func DefaultOptions() *Options {
	return &Options{ // Default options
		DetectTypes: nil,
		AstOptions: AstOptions{
			StyleCheck:      true,
			CheckIs:         true,
			DeepIsCheck:     false,
			UncheckedAssert: false,
			CheckUnused:     false,
		},
		Suggest: "",
	}
}
