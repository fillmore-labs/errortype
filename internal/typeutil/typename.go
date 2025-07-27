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

package typeutil

import (
	"bytes"
	"go/types"
	"strings"
)

// TypeName represents the fully qualified name of a Go type,
// consisting of its package path and its name.
// It is a simplified representation of a [types.TypeName] without type parameters.
type TypeName struct {
	Path string
	Name string
}

// NewTypeName creates a new TypeName from a [types.TypeName].
// It extracts the package path and the type's name.
// If the type is from [types.Universe], the Path will be empty.
func NewTypeName(tn *types.TypeName) TypeName {
	name := TypeName{
		Name: tn.Name(),
	}
	if pkg := tn.Pkg(); pkg != nil {
		name.Path = pkg.Path()
	}

	return name
}

// String returns the fully qualified name of the type ("pkg/path.TypeName").
// If the type has no package path, it returns just the type name.
func (t TypeName) String() string {
	if t.Path == "" {
		return t.Name
	}

	return t.Path + "." + t.Name
}

// Compare compares two [TypeName] instances lexicographically.
// It first compares by Path, and if they are equal, it compares by Name.
// It returns -1, 0, or 1.
func (t TypeName) Compare(other TypeName) int {
	if c := strings.Compare(t.Path, other.Path); c != 0 {
		return c
	}

	return strings.Compare(t.Name, other.Name)
}

// MarshalText implements encoding.TextMarshaler.
func (t TypeName) MarshalText() ([]byte, error) {
	var buf bytes.Buffer
	if t.Path != "" {
		buf.WriteString(t.Path)
		buf.WriteByte('.')
	}

	buf.WriteString(t.Name)

	return buf.Bytes(), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (t *TypeName) UnmarshalText(text []byte) error {
	if lastDotIndex := bytes.LastIndexByte(text, '.'); lastDotIndex >= 0 {
		t.Path = string(text[:lastDotIndex])
		t.Name = string(text[lastDotIndex+1:])

		return nil
	}

	t.Path = ""
	t.Name = string(text)

	return nil
}
