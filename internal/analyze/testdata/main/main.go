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

package main

import (
	"errors"
	"fmt"
)

type PointerError struct {
	Msg string
}

func (p *PointerError) Error() string {
	return p.Msg
}

func (p *PointerError) String() string {
	return p.Msg
}

type ValueError struct {
	Msg string
}

func (v ValueError) Error() string {
	return v.Msg
}

func (v ValueError) String() string {
	return v.Msg
}

type BadValueError struct {
	Msg string
}

func (v BadValueError) Error() string {
	return v.Msg
}

func (v BadValueError) String() string {
	return v.Msg
}

func ReturnPointer() error {
	return &PointerError{Msg: "pointer error"}
}

func ReturnValue() error {
	return ValueError{Msg: "value error"}
}

var (
	perr  = func() error { return &PointerError{Msg: "pointer error"} }()
	verr  = func() error { return ValueError{Msg: "value error"} }()
	pverr = func() error { return &BadValueError{Msg: "pointer to value error"} }() // want " \\(et:ret\\)$"
)

func main() {
	assert()

	errorsAs()

	typeSwitch()

	embedded()

	iface()
}

func assert() {
	var ok bool

	_, ok = perr.(*PointerError) // true
	fmt.Println(perr, ok)

	_, ok = verr.(ValueError) // true
	fmt.Println(verr, ok)

	_, ok = verr.(*ValueError) // want " \\(et:ast\\)$"
	fmt.Println("*", verr, ok)

	_, ok = pverr.(BadValueError) // false
	fmt.Println(pverr, ok)

	_, ok = pverr.(*BadValueError) // want " \\(et:ast\\)$"
	fmt.Println("*", pverr, ok)
}

func errorsAs() {
	var ok bool

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("panic:", r)
			}
		}()

		var p PointerError

		_ = errors.As(perr, &p) // want " \\(et:arg\\)$"
		panic("unreachable")
	}()

	var pp *PointerError

	ok = errors.As(perr, &pp) // true
	fmt.Println("As*", perr, ok)

	var v ValueError

	ok = errors.As(verr, &v) // true
	fmt.Println("As", verr, ok)

	var pv *ValueError

	ok = errors.As(verr, &pv) // want " \\(et:err\\)$"
	fmt.Println("As*", verr, ok)

	var bv BadValueError

	ok = errors.As(pverr, &bv) // false
	fmt.Println("As", pverr, ok)

	var pbv *BadValueError

	ok = errors.As(pverr, &pbv) // want " \\(et:err\\)$"
	fmt.Println("As*", pverr, ok)

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("panic:", r)
			}
		}()

		_ = errors.As(perr, &struct{}{}) // want " \\(et:arg\\)$"
		panic("unreachable")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("panic:", r)
			}
		}()

		_ = errors.As(perr, nil) // want " \\(et:arg\\)$"
		panic("unreachable")
	}()
}

func typeSwitch() {
	switch err := perr; e := err.(type) {
	case *PointerError:
		fmt.Println(e)

	default:
		panic(perr)
	}

	switch err := verr; e := err.(type) {
	case ValueError:
		fmt.Println(e)

	case *ValueError: // want " \\(et:ast\\)$"
		panic(verr)

	case nil:
		panic(verr)

	case any:
		panic(verr)

	default:
		panic(verr)
	}

	switch err := pverr; e := err.(type) {
	case BadValueError:
		panic(pverr)

	case *BadValueError: // want " \\(et:ast\\)$"
		fmt.Println(e)

	default:
		panic(pverr)
	}
}

func iface() {
	var pe interface {
		fmt.Stringer
		error
	} = &PointerError{Msg: "iface pointer"}

	var ve interface {
		fmt.Stringer
		error
	} = ValueError{Msg: "iface value"}

	var pve interface {
		fmt.Stringer
		error
	} = &BadValueError{Msg: "iface pointer to value"}

	_ = BadValueError{}

	type testT = fmt.Stringer

	var ok bool

	_, ok = pe.(testT) // true
	fmt.Println(pe, ok)

	_, ok = ve.(testT) // true
	fmt.Println(ve, ok)

	_, ok = pve.(testT) // true
	fmt.Println(pve, ok)

	var mye testT

	ok = errors.As(pe, &mye) // true
	fmt.Println("As", pe, ok)

	ok = errors.As(ve, &mye) // true
	fmt.Println("As", ve, ok)

	ok = errors.As(pve, &mye) // true
	fmt.Println("As", pve, ok)

	var myp any = &mye

	ok = errors.As(pe, myp) // true
	fmt.Println("As*", pe, ok)

	ok = errors.As(ve, myp) // true
	fmt.Println("As*", ve, ok)

	ok = errors.As(pve, myp) // true
	fmt.Println("As*", pve, ok)
}

type EmbeddedPointer struct{ *PointerError }

func embedded() {
	var eperr error = EmbeddedPointer{&PointerError{Msg: "embedded pointer"}}

	var _ error = &EmbeddedPointer{}

	var ok bool

	_, ok = eperr.(*EmbeddedPointer) // want " \\(et:emb\\+\\)$"
	fmt.Println(eperr, ok)

	_, ok = eperr.(EmbeddedPointer) // want " \\(et:emb\\)$"
	fmt.Println(eperr, ok)

	var ep EmbeddedPointer
	ok = errors.As(eperr, &ep) // want " \\(et:emb\\)$"
	fmt.Println("As", eperr, ok)

	var epp *EmbeddedPointer
	ok = errors.As(eperr, &epp) // want " \\(et:emb\\+\\)$"
	fmt.Println("As*", eperr, ok)

	var ep2err error = struct{ *PointerError }{&PointerError{Msg: "embedded pointer 2"}}

	var ep2 struct{ *PointerError }

	ok = errors.As(ep2err, &ep2) // true
	fmt.Println("As", ep2err, ok)

	var pep2err error = &struct{ *PointerError }{&PointerError{Msg: "embedded pointer 2"}}

	var pep2 *struct{ *PointerError }

	ok = errors.As(pep2err, &pep2) // true
	fmt.Println("As*", ep2err, ok)
}
