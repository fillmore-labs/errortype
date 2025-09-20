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
	"io"
	"slices"

	"github.com/goccy/go-yaml"

	"fillmore-labs.com/errortype/internal/errortypes"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// Write serializes the provided overrides suggestions into YAML format and writes it to the given io.Writer.
func Write(w io.Writer, suggestions map[errortypes.ErrorType][]typeutil.TypeName, name string) error {
	var suggestfile errorfileType

	for errortype, s := range suggestions {
		slices.SortFunc(s, typeutil.TypeName.Compare)

		switch errortype {
		case errortypes.PointerType:
			suggestfile.Pointer = s

		case errortypes.ValueType:
			suggestfile.Value = s

		case errortypes.Undecided:
			suggestfile.Inconsistent = s
		}
	}

	_, _ = w.Write([]byte("---"))
	if name != "" {
		_, _ = w.Write([]byte(" # suggestions for "))
		_, _ = w.Write([]byte(name))
	}
	_, _ = w.Write([]byte("\n"))

	return yaml.NewEncoder(w, yaml.IndentSequence(true)).Encode(suggestfile)
}
