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
	"io"
	"slices"

	"github.com/goccy/go-yaml"

	"fillmore-labs.com/errortype/facts"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// Write serializes the provided overrides suggestions into YAML format and writes it to the given io.Writer.
func Write(ctx context.Context, w io.Writer, suggestions Overrides, pkgPath string) error {
	var suggestFile errorFileType

	for errortype, s := range suggestions {
		slices.SortFunc(s, typeutil.TypeName.Compare)

		switch errortype {
		case facts.PointerType:
			suggestFile.Pointer = s

		case facts.ValueType:
			suggestFile.Value = s

		case facts.SuppressType:
			suggestFile.Inconsistent = s
		}
	}

	if _, err := io.WriteString(w, "---\n"); err != nil {
		return err
	}

	if pkgPath != "" {
		if _, err := io.WriteString(w, "# suggestions for "); err != nil {
			return err
		}

		if _, err := io.WriteString(w, pkgPath); err != nil {
			return err
		}

		if _, err := io.WriteString(w, "\n"); err != nil {
			return err
		}
	}

	enc := yaml.NewEncoder(w, yaml.IndentSequence(true))
	defer enc.Close()

	return enc.EncodeContext(ctx, suggestFile)
}
