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

package a

import "errors"

type convErrHolder struct {
	err error
	msg string
}

func consumeConv(_, _ error)        {}
func consumeOneConv(_ int, _ error) {}

func convErrPairOK() (ValueError, *my1Error) {
	return ValueError{}, &my1Error{}
}

func convErrPairBad() (*ValueError, my1Error) {
	return &ValueError{}, my1Error{}
}

func TestConversions(errCh chan error, errs []error) {
	// Regular and variadic function call arguments
	consumeConv(ValueError{}, &my1Error{})
	consumeConv(&ValueError{}, my1Error{}) // want ` \(et:arg\)$` ` \(et:arg\+\)$`
	consumeOneConv(1, &ValueError{})       // want ` \(et:arg\)$`

	_ = errors.Join(ValueError{}, &my1Error{})
	_ = errors.Join(&ValueError{}, my1Error{}) // want ` \(et:arg\)$` ` \(et:arg\+\)$`
	_ = errors.Join(errs...)

	// Explicit interface conversions
	_ = error(ValueError{})
	_ = error(&my1Error{})
	_ = error(&ValueError{}) // want ` \(et:arg\)$`
	_ = error(my1Error{})    // want ` \(et:arg\+\)$`

	// Multivalued tuple calls (fixed parameters and variadic)
	consumeConv(convErrPairOK())
	consumeConv(convErrPairBad()) // want ` \(et:arg\)$` ` \(et:arg\+\)$`
	_ = errors.Join(convErrPairOK())
	_ = errors.Join(convErrPairBad()) // want ` \(et:arg\)$` ` \(et:arg\+\)$`

	// Composite literals: slice, array, map, struct, and elided pointer elements
	_ = []error{ValueError{}, &my1Error{}}
	_ = []error{&ValueError{}, 2: my1Error{}} // want ` \(et:asn\)$` ` \(et:asn\+\)$`
	_ = [2]error{&ValueError{}, my1Error{}}   // want ` \(et:asn\)$` ` \(et:asn\+\)$`
	_ = map[error]error{ValueError{}: &my1Error{}}
	_ = map[error]error{&ValueError{}: my1Error{}} // want ` \(et:asn\)$` ` \(et:asn\+\)$`
	_ = map[string]error{"k": &ValueError{}}       // want ` \(et:asn\)$`

	_ = convErrHolder{err: ValueError{}, msg: "ok"}
	_ = convErrHolder{err: &ValueError{}} // want ` \(et:asn\)$`
	_ = convErrHolder{my1Error{}, "bad"}  // want ` \(et:asn\+\)$`
	_ = []*convErrHolder{
		{err: &ValueError{}}, // want ` \(et:asn\)$`
		{err: my1Error{}},    // want ` \(et:asn\+\)$`
	}

	// Channel sends
	errCh <- ValueError{}
	errCh <- &my1Error{}
	errCh <- &ValueError{} // want ` \(et:asn\)$`
	errCh <- my1Error{}    // want ` \(et:asn\+\)$`
}
