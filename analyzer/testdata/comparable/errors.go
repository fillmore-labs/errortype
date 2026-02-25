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

package comparable

import "errors"

type comparableError struct{ _ [0]byte }

func (comparableError) Error() string {
	return ""
}

type nonComparableError struct{ _ [0][]byte } // want ` \(et:nce\+\)$`

func (nonComparableError) Error() string {
	return ""
}

type nonComparableAliasError = nonComparableError // want ` \(et:nce\+\)$`

type nonComparableWithIsError struct{ _ [0][]byte }

func (nonComparableWithIsError) Error() string {
	return ""
}

func (nonComparableWithIsError) Is(err error) bool {
	return errors.As(err, new(nonComparableWithIsError)) // want "only shallowly compare"
}

func NonComparable() {
	_ = errors.Is(comparableError{}, comparableError{})

	_ = errors.Is(nonComparableError{}, nonComparableError{}) // want "not comparable .* is always false"

	_ = errors.Is(nonComparableWithIsError{}, nonComparableWithIsError{})
}

type UnnamedIsError struct{ _ int }

func (UnnamedIsError) Error() string { return "unnamed" }

func (UnnamedIsError) Is(error) bool {
	return false
}

type UnderscoreIsError struct{ _ int }

func (_ UnderscoreIsError) Error() string { return "underscore" }

func (_ UnderscoreIsError) Is(_ error) bool {
	return false
}

type nonComparablePointerIsError struct{ _ [0][]byte }

func (nonComparablePointerIsError) Error() string { return "" }

func (*nonComparablePointerIsError) Is(error) bool { return false }

var nonComparableSentinel = nonComparableError{} // want `Not comparable .* should implement.* "Is" method` ` \(et:nam\)$`

var nonComparableSentinel2 error = nonComparableError{} // want `Not comparable .* should implement.* "Is" method` ` \(et:nam\)$`

func Errors5() {
	_ = errors.Is(&UnderscoreIsError{}, &UnnamedIsError{})

	_ = errors.Is(&UnnamedIsError{}, &UnderscoreIsError{})

	_ = errors.Is(nonComparableError{}, UnderscoreIsError{}) // want "not comparable.* is always false"

	_ = errors.Is(UnderscoreIsError{}, nonComparableError{}) // want "not comparable.* is always false"

	_ = errors.Is(UnderscoreIsError{}, nonComparableSentinel) // want "not comparable.* is always false"

	_ = errors.Is(UnderscoreIsError{}, nonComparableSentinel2)
}
