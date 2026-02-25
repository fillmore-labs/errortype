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
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"

	"fillmore-labs.com/errortype/internal/detectwrappers/result"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// Wrappers associates a wrapper type with a list of fully qualified function or method names.
type Wrappers struct {
	Wrappers map[result.WrapperType][]typeutil.FuncName `yaml:"wrappers,omitempty"`
}

// ReadWrappers parses an override file from the provided io.Reader and returns a map
// associating wrapper types with their corresponding function names. The override file
// is expected to be in YAML format and structured according to errorFileTypeWrappers.
func ReadWrappers(ctx context.Context, r io.Reader) (Wrappers, error) {
	dec := yaml.NewDecoder(r)

	var errorfile Wrappers
	if err := dec.DecodeContext(ctx, &errorfile); err != nil {
		if errors.Is(err, io.EOF) {
			return Wrappers{}, nil
		}

		return Wrappers{}, fmt.Errorf("error parsing override file: %w", err)
	}

	return errorfile, nil
}

// ReadWrappersFile reads wrapper function overrides from the specified file.
func ReadWrappersFile(ctx context.Context, fileName string) (Wrappers, error) {
	overridesFile, err := os.Open(filepath.Clean(fileName))
	if err != nil {
		return Wrappers{}, err
	}

	defer overridesFile.Close() // ignore error

	return ReadWrappers(ctx, overridesFile)
}
