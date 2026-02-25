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

// Package forms checks how suggestions are written for different underlying types.
package forms

import "fmt"

// Pointer errors.

type StructError struct{ Code int }

func (StructError) Error() string { return "struct" }

var _ error = (*StructError)(nil)

type CodeError int

func (CodeError) Error() string { return "code" }

var _ error = (*CodeError)(nil)

func ReturnStruct() error { return StructError{} } // want `should be returned as a pointer \("&StructError\{\.\.\.\}"\), not by value\. \(et:ret\+\)$`

func ReturnCode() error { return CodeError(1) } // want `should be returned as a pointer \("new\(CodeError\(\.\.\.\)\)"\), not by value\. \(et:ret\+\)$`

func WrapCode() error { return fmt.Errorf("w: %w", CodeError(1)) } // want `should be wrapped as a pointer \("new\(CodeError\(\.\.\.\)\)"\), not a value\. \(et:wrp\+\)$`

var ErrCode error = CodeError(1) // want `"ErrCode" should be a pointer \("\.\.\. = new\(CodeError\(\.\.\.\)\)"\), not a value\. \(et:var\+\)$`

// Value errors.

type MsgError string

func (MsgError) Error() string { return "msg" }

var _ error = MsgError("")

type FlagError bool

func (FlagError) Error() string { return "flag" }

var _ error = FlagError(false)

type FuncError func() string

func (f FuncError) Error() string { return f() }

var _ error = FuncError(nil)

type ListError []string

func (ListError) Error() string { return "list" }

var _ error = ListError{}

func ReturnMsg() error { return new(MsgError("x")) } // want `should be returned by value \("MsgError\(\.\.\.\)"\), not as a pointer\. \(et:ret\)$`

func ReturnList() error { return &ListError{} } // want `should be returned by value \("ListError\{\.\.\.\}"\), not as a pointer\. \(et:ret\)$`

// Compile-time assertions.

type AssertCodeError int

func (AssertCodeError) Error() string { return "code" }

func (AssertCodeError) Timeout() bool { return false }

var (
	_ error = (*AssertCodeError)(nil) // want `Compile-time assertion should be a value \("var _ error = AssertCodeError\(0\)"\), not a pointer\. \(et:var\)$`
	_ error = AssertCodeError(0)
	_ error = AssertCodeError(1)
)
