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

	"fillmore-labs.com/errortype/detect/result"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// Read parses an override file from the provided io.Reader and returns a map
// associating type names with their corresponding error types. The override file
// is expected to be in YAML format and structured according to errorFileType.
func Read(ctx context.Context, r io.Reader) (Overrides, error) {
	dec := yaml.NewDecoder(r)

	var errorfile errorFileType
	if err := dec.DecodeContext(ctx, &errorfile); err != nil {
		if errors.Is(err, io.EOF) {
			return Overrides{}, nil
		}

		return Overrides{}, fmt.Errorf("error parsing override file: %w", err)
	}

	return Overrides{
			Types: map[result.ErrorType][]typeutil.TypeName{
				result.Pointer:  errorfile.Pointer,
				result.Value:    errorfile.Value,
				result.Suppress: errorfile.Suppress,
				// errortypes.InconsistentType are ignored.
			},
			Wrappers: errorfile.Wrappers,
		},
		nil
}

// ReadFile reads error type usage overrides from the specified file.
func ReadFile(ctx context.Context, fileName string) (Overrides, error) {
	overridesFile, err := os.Open(filepath.Clean(fileName))
	if err != nil {
		return Overrides{}, fmt.Errorf("can't open overrides file: %w", err)
	}

	defer overridesFile.Close() // ignore error

	return Read(ctx, overridesFile)
}
