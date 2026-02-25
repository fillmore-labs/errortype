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
	"errors"
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
	WrapperErrorf
	FuncIsType
	FuncEqual
	FuncAssert

	LastWrappperType
)

var wrapperTypes = [...]string{
	WrapperUnknown: "unknown",
	WrapperIs:      "is",
	WrapperAs:      "as",
	WrapperAsType:  "astype",
	WrapperErrorf:  "errorf",
	FuncIsType:     "istype",
	FuncEqual:      "equal",
	FuncAssert:     "assert",
}

func (w WrapperType) String() string {
	if w >= LastWrappperType {
		return "WrapperType(" + strconv.FormatInt(int64(w), 10) + ")"
	}

	return wrapperTypes[w]
}

// UnmarshalText implements [encoding.TextUnmarshaler].
func (w *WrapperType) UnmarshalText(text []byte) error {
	s := string(text)

	for i, t := range wrapperTypes {
		if i == 0 || s != t {
			continue
		}

		*w = WrapperType(i)

		return nil
	}

	return fmt.Errorf("unknown wrapper type: %q", s)
}

// MarshalText implements [encoding.TextMarshaler].
func (w WrapperType) MarshalText() ([]byte, error) {
	return w.AppendText(nil)
}

// AppendText implements [encoding.TextAppender].
func (w WrapperType) AppendText(buf []byte) ([]byte, error) {
	if w == WrapperUnknown || w >= LastWrappperType {
		return nil, errors.New("cannot serialize " + w.String())
	}

	buf = append(buf, wrapperTypes[w]...)

	return buf, nil
}
