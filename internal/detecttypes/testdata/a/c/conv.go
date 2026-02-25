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

package c

type (
	// Suggestion 1: General conversions to error
	AssignPtrError      struct{ error } // want AssignPtrError:"pointer"
	AssignValError      struct{ error } // want AssignValError:"value"
	MultiAssignError    struct{ error } // want MultiAssignError:"pointer"
	CallArgPtrError     struct{ error } // want CallArgPtrError:"pointer"
	VariadicArgValError struct{ error } // want VariadicArgValError:"value"
	MultiCallError      struct{ error } // want MultiCallError:"pointer"
	ChanSendPtrError    struct{ error } // want ChanSendPtrError:"pointer"
	ExplicitConvError   struct{ error } // want ExplicitConvError:"pointer"
	SliceElemPtrError   struct{ error } // want SliceElemPtrError:"pointer"
	MapElemValError     struct{ error } // want MapElemValError:"value"
	StructFieldPtrError struct{ error } // want StructFieldPtrError:"pointer"
	StructPosValError   struct{ error } // want StructPosValError:"value"

	// Suggestion 2: Multi-error returns, multi-valued returns, concrete return signatures
	FirstReturnPtrError struct{ error } // want FirstReturnPtrError:"pointer"
	MultiReturnPtrError struct{ error } // want MultiReturnPtrError:"pointer"
	ConcreteSigPtrError struct{ error } // want ConcreteSigPtrError:"pointer"
	ConstructorPtrError struct{ error } // want ConstructorPtrError:"pointer"

	// Suggestion 3: Type switches, nested assertions, generic constraints
	TypeSwitchPtrError  struct{ error } // want TypeSwitchPtrError:"pointer"
	TypeSwitchValError  struct{ error } // want TypeSwitchValError:"value"
	CondAssertPtrError  struct{ error } // want CondAssertPtrError:"pointer"
	GenericBoxPtrError  struct{ error } // want GenericBoxPtrError:"pointer"
	GenericFuncValError struct{ error } // want GenericFuncValError:"value"

	// Suggestion 4: Context-sensitive literals, &e, nested/elided literals
	MutatedLocalPtrError struct{ error } // want MutatedLocalPtrError:"pointer"
	NestedSlicePtrError  struct{ error } // want NestedSlicePtrError:"pointer"
	Elided2DPtrError     struct{ error } // want Elided2DPtrError:"pointer"
)

type MutatedCastPtrError int // want MutatedCastPtrError:"pointer"

func (MutatedCastPtrError) Error() string { return "" }

type errBox[E error] struct{ err E }

func acceptErr(_ error)                        {}
func acceptVariadic(_ ...error)                {}
func acceptTwo(_ int, _ error)                 {}
func multiAssign() (int, *MultiAssignError)    { return 0, nil }
func multiCall() (int, *MultiCallError)        { return 0, nil }
func multiReturn() (int, *MultiReturnPtrError) { return 0, nil }
func genericErrFunc[E error](_ E)              {}

func testConversions(
	ap *AssignPtrError,
	av AssignValError,
	cp *CallArgPtrError,
	vv VariadicArgValError,
	csp *ChanSendPtrError,
	ecp *ExplicitConvError,
	sep *SliceElemPtrError,
	mev MapElemValError,
	sfp *StructFieldPtrError,
	spv StructPosValError,
	gfv GenericFuncValError,
	ch chan<- error,
) {
	var err error
	err = ap
	err = av

	var n int
	n, err = multiAssign()
	_ = n

	acceptErr(cp)
	acceptVariadic(vv)
	acceptTwo(multiCall())

	ch <- csp
	_ = error(ecp)

	_ = []error{sep}
	_ = map[string]error{"k": mev}
	_ = struct{ e error }{e: sfp}
	_ = struct{ e error }{spv}

	var _ errBox[*GenericBoxPtrError]
	genericErrFunc(gfv)
	_ = err
}

func testFirstReturn(p *FirstReturnPtrError) (error, bool) {
	return p, true
}

func testSpreadReturn() (int, error) {
	return multiReturn()
}

func testConcreteSig() *ConcreteSigPtrError {
	return nil
}

func testConstructor() (*ConstructorPtrError, bool) {
	e := ConstructorPtrError{}
	return &e, true
}

func testTypeSwitchAndCondAssert(err error) {
	switch err.(type) {
	case *TypeSwitchPtrError:
	case TypeSwitchValError:
	}

	if err != nil && err.(*CondAssertPtrError) != nil {
		return
	}
}

func testContextSensitive() {
	var err error

	e1 := MutatedLocalPtrError{}
	err = &e1

	e2 := MutatedCastPtrError(1)
	err = &e2

	_ = []error{&NestedSlicePtrError{}}
	_ = [][]*Elided2DPtrError{{{}}}

	_ = err
}
