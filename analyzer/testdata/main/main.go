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
	errp  = func() error { return &PointerError{Msg: "pointer error"} }()
	errv  = func() error { return ValueError{Msg: "value error"} }()
	errpv = func() error { return &BadValueError{Msg: "pointer to value error"} }() // want ` \(et:ret\)$`
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

	_, ok = errp.(*PointerError) // true
	fmt.Println(errp, ok)

	_, ok = errv.(ValueError) // true
	fmt.Println(errv, ok)

	_, ok = errv.(*ValueError) // want ` \(et:ast\)$`
	fmt.Println("*", errv, ok)

	_, ok = errpv.(BadValueError) // false
	fmt.Println(errpv, ok)

	_, ok = errpv.(*BadValueError) // want ` \(et:ast\)$`
	fmt.Println("*", errpv, ok)
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

		_ = errors.As(errp, &p) // want ` \(et:arg\)$`
		panic("unreachable")
	}()

	var pp *PointerError

	ok = errors.As(errp, &pp) // true
	fmt.Println("As*", errp, ok)

	var v ValueError

	ok = errors.As(errv, &v) // true
	fmt.Println("As", errv, ok)

	var pv *ValueError

	ok = errors.As(errv, &pv) // want ` \(et:err\)$`
	fmt.Println("As*", errv, ok)

	var bv BadValueError

	ok = errors.As(errpv, &bv) // false
	fmt.Println("As", errpv, ok)

	var pbv *BadValueError

	ok = errors.As(errpv, &pbv) // want ` \(et:err\)$`
	fmt.Println("As*", errpv, ok)

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("panic:", r)
			}
		}()

		_ = errors.As(errp, &struct{}{}) // want ` \(et:arg\)$`
		panic("unreachable")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("panic:", r)
			}
		}()

		_ = errors.As(errp, nil) // want ` \(et:arg\)$`
		panic("unreachable")
	}()
}

func typeSwitch() {
	switch err := errp; e := err.(type) {
	case *PointerError:
		fmt.Println(e)

	default:
		panic(errp)
	}

	switch err := errv; e := err.(type) {
	case ValueError:
		fmt.Println(e)

	case *ValueError: // want ` \(et:ast\)$`
		panic(errv)

	case nil:
		panic(errv)

	case any:
		panic(errv)

	default:
		panic(errv)
	}

	switch err := errpv; e := err.(type) {
	case BadValueError:
		panic(errpv)

	case *BadValueError: // want ` \(et:ast\)$`
		fmt.Println(e)

	default:
		panic(errpv)
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

	ok = errors.As(pe, myp)
	fmt.Println("As*", pe, ok)

	ok = errors.As(ve, myp)
	fmt.Println("As*", ve, ok)

	ok = errors.As(pve, myp)
	fmt.Println("As*", pve, ok)
}

type PointerEmbeddedError struct{ *PointerError }

func embedded() {
	var eperr error = PointerEmbeddedError{&PointerError{Msg: "embedded pointer"}}

	var _ error = &PointerEmbeddedError{} // want ` \(et:emb\+\)$`

	var ok bool

	_, ok = eperr.(*PointerEmbeddedError) // want ` \(et:emb\+\)$`
	fmt.Println(eperr, ok)

	_, ok = eperr.(PointerEmbeddedError) // want ` \(et:emb\)$`
	fmt.Println(eperr, ok)

	var ep PointerEmbeddedError
	ok = errors.As(eperr, &ep) // want ` \(et:emb\)$`
	fmt.Println("As", eperr, ok)

	var epp *PointerEmbeddedError
	ok = errors.As(eperr, &epp) // want ` \(et:emb\+\)$`
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
