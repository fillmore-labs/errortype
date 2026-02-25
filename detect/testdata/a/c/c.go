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

package c

type (
	ValueDefaultError    struct{}               // want ValueDefaultError:"value"
	ValueFuncError       struct{}               // want ValueFuncError:"value"
	ValueVarError        struct{}               // want ValueVarError:"value"
	PointerDefaultError  struct{}               // want PointerDefaultError:"pointer"
	PointerFuncError     struct{}               // want PointerFuncError:"pointer"
	PointerVarError      struct{}               // want PointerVarError:"pointer"
	EmbeddedDefaultError struct{ error }        // want EmbeddedDefaultError:"undecided"
	EmbeddedFuncError    struct{ error }        // want EmbeddedFuncError:"pointer"
	EmbeddedVarError     struct{ error }        // want EmbeddedVarError:"value"
	AliasError           = ValueDefaultError    // want AliasError:"value"
	PointerAliasError    = *PointerDefaultError // want PointerAliasError:"value"
)

func (ValueDefaultError) Error() string { return "" } // value type
func (ValueFuncError) Error() string    { return "" } // overwritten by func
func (ValueVarError) Error() string     { return "" } // overwritten by var

func (*PointerDefaultError) Error() string { return "" } // pointer type
func (PointerFuncError) Error() string     { return "" } // overwritten by func
func (PointerVarError) Error() string      { return "" } // overwritten by var

func NewValueFunc() error { return ValueFuncError{} } // value type
func NewValueVar() error  { return &ValueVarError{} } // overwritten by var

func NewPointerFunc() error { return &PointerFuncError{} } // pointer type
func NewPointerVar() error  { return PointerVarError{} }   // overwritten by var

func NewEmbeddedFunc() error { return &EmbeddedFuncError{} } // pointer type
func NewEmbeddedVar() error  { return &EmbeddedVarError{} }  // overwritten by var

func NewEmbeddedDefault1() error { return EmbeddedDefaultError{} }  // contradictory, ignored
func NewEmbeddedDefault2() error { return &EmbeddedDefaultError{} } // contradictory, ignored

func NewPointerDefault() any { return PointerDefaultError{} } // ignored, doesn't implement error

func Ignored() error {
	err := func() error { return ValueDefaultError{} }()

	return err
}

var (
	_ error = ValueVarError{} // value type

	_ error = (*PointerVarError)(nil) // pointer type

	_ error = EmbeddedVarError{} // value type

	_, _ error = EmbeddedDefaultError{}, (*EmbeddedDefaultError)(nil) // contradictory, ignored

	// _ error = PointerDefault{} // type error.
)
