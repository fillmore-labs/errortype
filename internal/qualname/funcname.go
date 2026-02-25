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

package qualname

import (
	"bytes"
	"errors"
	"fmt"
	"go/types"
	"slices"
	"strings"
)

// Func represents the fully qualified name of a function, method, or variable.
// It deconstructs an identity into its constituent parts: package path,
// receiver type, and name, ignoring type parameters.
type Func struct {
	// Path is the package path ("encoding/json").
	Path string

	// LocalFunc is the package-local part.
	LocalFunc
}

// LocalFunc is a package-local name of a function, method, or variable.
type LocalFunc struct {
	// Receiver is the name of the receiver type ("Decoder").
	// It is empty for regular functions and variables.
	Receiver string

	// Name is the function, method, or variable name.
	Name string
}

// FuncOf extracts the name components of a given types.Object
// (which can be a *types.Func or a *types.Var).
// It populates a Func struct, which is simplified and canonicalized,
// and can then be used as a map index or to get a string representation.
func FuncOf(obj types.Object) Func {
	var f Func
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
func (f Func) String() string {
	txt, _ := f.MarshalText()
	return string(txt)
}

// String returns the local function name as a string.
// For a method, the format is "(<receiver>).<name>".
// For a function, the format is <name>".
func (l LocalFunc) String() string {
	return Func{LocalFunc: l}.String()
}

// Compare compares two [Func] instances lexicographically.
// It first compares by Path, and if they are equal, it compares by Receiver/Name.
// It returns -1, 0, or 1.
func (f Func) Compare(other Func) int {
	if c := strings.Compare(f.Path, other.Path); c != 0 {
		return c
	}

	return f.LocalFunc.Compare(other.LocalFunc)
}

// Compare two [LocalFunc] instances lexicographically by Receiver/Name.
func (l LocalFunc) Compare(other LocalFunc) int {
	if c := strings.Compare(l.Receiver, other.Receiver); c != 0 {
		return c
	}

	return strings.Compare(l.Name, other.Name)
}

// MarshalText implements [encoding.TextMarshaler].
func (f Func) MarshalText() ([]byte, error) {
	return f.AppendText(nil)
}

// AppendText implements [encoding.TextAppender].
func (f Func) AppendText(buf []byte) ([]byte, error) {
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
func (f *Func) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return fmt.Errorf("%w: %q", errInvalidFuncName, text)
	}

	var path, receiver, name []byte

	if text[0] == '(' {
		// It's a method
		closeParen := bytes.IndexByte(text, ')')
		if closeParen < 0 || closeParen+1 >= len(text) || text[closeParen+1] != '.' {
			return fmt.Errorf("%w: %q", errInvalidFuncName, text)
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
			return fmt.Errorf("%w: %q", errInvalidFuncName, text)
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
		return fmt.Errorf("%w: %q", errInvalidFuncName, text)
	}

	*f = Func{
		Path: string(path),
		LocalFunc: LocalFunc{
			Receiver: string(receiver),
			Name:     string(name),
		},
	}

	return nil
}
