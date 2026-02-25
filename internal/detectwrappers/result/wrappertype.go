// Copyright 2026 Oliver Eikemeier. All Rights Reserved.
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

package result

import (
	"fmt"
	"strconv"
)

// ErrorFunc describes a detected wrapper function around [errors.Is], [errors.As], or [errors.AsType].
type ErrorFunc struct {
	Type   WrapperType
	Source int8 // Index of the source error argument
	Target int8 // Index of the target argument (Is/As) or type parameter (AsType)
}

// AFact makes *ErrorFunc satisfy the [analysis.Fact] interface.
func (*ErrorFunc) AFact() {}

func (w ErrorFunc) String() string {
	return fmt.Sprintf("%s(%d, %d)", w.Type, w.Source, w.Target)
}

// WrapperType identifies the type of error wrapper function.
type WrapperType uint8

// Supported wrapper types.
const (
	WrapperUnknown WrapperType = iota
	WrapperIs
	WrapperAs
	WrapperAsType
	WrapperUnwrap
	WrapperErrorf
	FuncIsType
	FuncEqual
	FuncAssert

	NumWrappperTypes = iota
)

var wrapperTypes = [...]string{
	WrapperUnknown: "unknown",
	WrapperIs:      "is",
	WrapperAs:      "as",
	WrapperAsType:  "astype",
	WrapperUnwrap:  "unwrap",
	WrapperErrorf:  "errorf",
	FuncIsType:     "istype",
	FuncEqual:      "equal",
	FuncAssert:     "assert",
}

func (w WrapperType) String() string {
	if w >= NumWrappperTypes {
		return "WrapperType(" + strconv.FormatInt(int64(w), 10) + ")"
	}

	return wrapperTypes[w]
}

// configurableTypes are the wrapper types that can be configured as explicit wrappers.
// The text form is only used in configuration, so [WrapperType.UnmarshalText] accepts
// just these. Expand this list deliberately; every entry becomes part of the configuration format.
var configurableTypes = [...]WrapperType{WrapperIs, WrapperAs, WrapperAsType, WrapperErrorf}

// UnmarshalText implements [encoding.TextUnmarshaler].
//
// It accepts only wrapper types that can be configured: "is", "as", "astype" and "errorf".
func (w *WrapperType) UnmarshalText(text []byte) error {
	s := string(text)

	for _, t := range configurableTypes {
		if s != wrapperTypes[t] {
			continue
		}

		*w = t

		return nil
	}

	return fmt.Errorf(`unknown wrapper type: %q, expected "is", "as", "astype" or "errorf"`, s)
}

// MarshalText implements [encoding.TextMarshaler].
func (w WrapperType) MarshalText() ([]byte, error) {
	return w.AppendText(nil)
}

// AppendText implements [encoding.TextAppender].
func (w WrapperType) AppendText(buf []byte) ([]byte, error) {
	if w == WrapperUnknown || w >= NumWrappperTypes {
		return nil, SerializationError(w)
	}

	buf = append(buf, wrapperTypes[w]...)

	return buf, nil
}

// SerializationError indicates that a WrapperType cannot be serialized to text.
type SerializationError WrapperType

func (e SerializationError) Error() string {
	return "cannot serialize " + WrapperType(e).String()
}
