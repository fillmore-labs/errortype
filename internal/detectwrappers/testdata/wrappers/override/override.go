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

package override

func Is(err, target error) bool { // want Is:`is\(0, 1\)`
	return false
}

func As(err error, target any) bool { // want As:`as\(0, 1\)`
	return false
}

func AsType[E error](err error) (E, bool) { // want AsType:`astype\(0, 0\)`
	return *new(E), false
}

func Errorf(format string, a ...any) error { // want Errorf:`errorf\(0, 1\)`
	return nil
}

func WrapsIs(err, target error) bool { // want WrapsIs:`is\(0, 1\)`
	return Is(err, target)
}

type suite struct{}

func (suite) Is(err, target error) bool { // want Is:`is\(0, 1\)`
	return false
}

func (suite) As(err error, target any) bool { // want As:`as\(0, 1\)`
	return false
}

func (suite) Errorf(format string, a ...any) error { // want Errorf:`errorf\(0, 1\)`
	return nil
}
