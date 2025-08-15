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
	"errors"
	"fmt"
	"io"

	"github.com/goccy/go-yaml"

	"fillmore-labs.com/errortype/internal/errortypes"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// Read parses an override file from the provided io.Reader and returns a map
// associating type names with their corresponding error types. The override file
// is expected to be in YAML format and structured according to errorfileType.
func Read(r io.Reader) ([]Override, error) {
	dec := yaml.NewDecoder(r)

	var errorfile errorfileType
	if err := dec.Decode(&errorfile); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, nil
		}

		return nil, fmt.Errorf("error parsing override file: %w", err)
	}

	errorfileMap := map[errortypes.ErrorType][]typeutil.TypeName{
		errortypes.PointerType:  errorfile.Pointer,
		errortypes.ValueType:    errorfile.Value,
		errortypes.SuppressType: errorfile.Suppress,
		// errortypes.InconsistentType are ignored.
	}

	var overrides []Override

	for override, types := range errorfileMap {
		for _, typeName := range types {
			overrides = append(overrides, Override{TypeName: typeName, ErrorType: override})
		}
	}

	return overrides, nil
}
