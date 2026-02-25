// Copyright 2025-2026 Oliver Eikemeier. All Rights Reserved.
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
	"errors"
	"fmt"
	"go/types"
	"slices"
	"strings"
)

// FuncName represents the fully qualified name of a function, method, or variable.
// It deconstructs an identity into its constituent parts: package path,
// receiver type, and name, ignoring type parameters.
type FuncName struct {
	// Path is the package path ("encoding/json").
	Path string

	// LocalFuncName is the package-local part.
	LocalFuncName
}

// LocalFuncName is a package-local name of a function, method, or variable.
type LocalFuncName struct {
	// Receiver is the name of the receiver type ("Decoder").
	// It is empty for regular functions and variables.
	Receiver string

	// Name is the function, method, or variable name.
	Name string
}

// FuncNameOf extracts the name components of a given types.Object
// (which can be a *types.Func or a *types.Var).
// It populates a FuncName struct, which is simplified and canonicalized,
// and can then be used as a map index or to get a string representation.
func FuncNameOf(obj types.Object) FuncName {
	var f FuncName
	f.Name = obj.Name()

	var recv *types.Var
	if fun, ok := obj.(*types.Func); ok {
		recv = fun.Signature().Recv()
	}

	if recv == nil { // It's a regular function.
		if pkg := obj.Pkg(); pkg != nil {
			f.Path = pkg.Path()
		}

		return f
	}

	rtyp := recv.Type() // It's a method.

recvloop:
	switch t := rtyp.(type) {
	case *types.Alias:
		rtyp = t.Rhs() // Unwrap alias.
		goto recvloop

	case *types.Pointer:
		rtyp = t.Elem() // If it's a pointer, unwrap to the element type.
		goto recvloop

	case *types.Named:
		tn := t.Obj()
		if pkg := tn.Pkg(); pkg != nil {
			f.Path = pkg.Path()
		}
		f.Receiver = tn.Name()

	case *types.Interface: // Method on an interface type.
		f.Receiver = "interface"

	default: // Anonymous types shouldn't have methods.
		f.Receiver = "<invalid>"
	}

	return f
}

// String returns the fully qualified function name as a string.
// For a method, the format is "(<path>.<receiver>).<name>".
// For a function, the format is "<path>.<name>".
func (f FuncName) String() string {
	txt, _ := f.MarshalText()
	return string(txt)
}

// String returns the local function name as a string.
// For a method, the format is "(<receiver>).<name>".
// For a function, the format is <name>".
func (l LocalFuncName) String() string {
	return FuncName{LocalFuncName: l}.String()
}

// Compare compares two [FuncName] instances lexicographically.
// It first compares by Path, and if they are equal, it compares by Receiver/Name.
// It returns -1, 0, or 1.
func (f FuncName) Compare(other FuncName) int {
	if c := strings.Compare(f.Path, other.Path); c != 0 {
		return c
	}

	return f.LocalFuncName.Compare(other.LocalFuncName)
}

// Compare two [LocalFuncName] instances lexicographically by Receiver/Name.
func (l LocalFuncName) Compare(other LocalFuncName) int {
	if c := strings.Compare(l.Receiver, other.Receiver); c != 0 {
		return c
	}

	return strings.Compare(l.Name, other.Name)
}

// MarshalText implements [encoding.TextMarshaler].
func (f FuncName) MarshalText() ([]byte, error) {
	return f.AppendText(nil)
}

// AppendText implements [encoding.TextAppender].
func (f FuncName) AppendText(buf []byte) ([]byte, error) {
	plen := len(f.Path)
	rlen := len(f.Receiver)

	size := len(f.Name)
	if plen > 0 {
		size += plen + 1
	}

	if rlen > 0 {
		size += rlen + 3
	}

	buf = slices.Grow(buf, size)

	if rlen > 0 {
		buf = append(buf, '(')
	}

	if plen > 0 {
		buf = append(buf, f.Path...)
		buf = append(buf, '.')
	}

	if rlen > 0 {
		buf = append(buf, f.Receiver...)
		buf = append(buf, ')', '.')
	}

	buf = append(buf, f.Name...)

	return buf, nil
}

var errInvalidFuncName = errors.New("invalid function name")

// UnmarshalText implements [encoding.TextUnmarshaler].
func (f *FuncName) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return fmt.Errorf("%w %q", errInvalidFuncName, text)
	}

	var path, receiver, name []byte

	if text[0] == '(' {
		// It's a method
		closeParen := bytes.IndexByte(text, ')')
		if closeParen < 0 || closeParen+1 >= len(text) || text[closeParen+1] != '.' {
			return fmt.Errorf("%w %q", errInvalidFuncName, text)
		}

		name = text[closeParen+2:]

		receiverPart := text[1:closeParen]
		if len(receiverPart) > 0 && receiverPart[0] == '*' {
			receiverPart = receiverPart[1:]
		}

		if lastDot := bytes.LastIndexByte(receiverPart, '.'); lastDot < 0 {
			receiver = receiverPart
		} else {
			path = receiverPart[:lastDot]
			receiver = receiverPart[lastDot+1:]
		}

		if len(receiver) == 0 {
			return fmt.Errorf("%w %q", errInvalidFuncName, text)
		}
	} else {
		// It's a regular function
		if lastDot := bytes.LastIndexByte(text, '.'); lastDot < 0 {
			name = text
		} else {
			path = text[:lastDot]
			name = text[lastDot+1:]
		}
	}

	if len(name) == 0 {
		return fmt.Errorf("%w %q", errInvalidFuncName, text)
	}

	*f = FuncName{
		Path: string(path),
		LocalFuncName: LocalFuncName{
			Receiver: string(receiver),
			Name:     string(name),
		},
	}

	return nil
}
