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
	ValueDefault    struct{}
	ValueFunc       struct{}
	ValueVar        struct{}
	PointerDefault  struct{}
	PointerFunc     struct{}
	PointerVar      struct{}
	EmbeddedDefault struct{ error }
	EmbeddedFunc    struct{ error }
	EmbeddedVar     struct{ error }

	LocalOverride struct{ error }

	Alias = ValueDefault
)

func (ValueDefault) Error() string { return "" } // value type
func (ValueFunc) Error() string    { return "" } // overwritten by func
func (ValueVar) Error() string     { return "" } // overwritten by var

func (*PointerDefault) Error() string { return "" } // pointer type
func (PointerFunc) Error() string     { return "" } // overwritten by func
func (PointerVar) Error() string      { return "" } // overwritten by var

func NewValueFunc() error { return ValueFunc{} } // value type
func NewValueVar() error  { return &ValueVar{} } // overwritten by var

func NewPointerFunc() error { return &PointerFunc{} } // pointer type
func NewPointerVar() error  { return PointerVar{} }   // overwritten by var

func NewEmbeddedFunc() error { return &EmbeddedFunc{} } // pointer type
func NewEmbeddedVar() error  { return &EmbeddedVar{} }  // overwritten by var

func NewEmbeddedDefault1() error { return EmbeddedDefault{} }  // contradictory, ignored
func NewEmbeddedDefault2() error { return &EmbeddedDefault{} } // contradictory, ignored

func NewLocalOverride() error { return &LocalOverride{} } // overwritten by local var

func NewPointerDefault() any { return PointerDefault{} } // ignored, doesn't implement error

func Ignored() error {
	err := func() error { return ValueDefault{} }()

	return err
}

var (
	_ error = ValueVar{} // value type

	_ error = (*PointerVar)(nil) // pointer type

	_ error = EmbeddedVar{} // value type

	_, _ error = EmbeddedDefault{}, (*EmbeddedDefault)(nil) // contradictory, ignored

	// _ error = PointerDefault{} // type error.
)
