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

package wrappers

import (
	"errors"
	"testing"
)

func myErrorAs(err error, target interface{}) bool { // want myErrorAs:`as\(0, 1\)`
	return errors.As(err, target)
}

type Suite struct{}

func (Suite) Fail() { panic("fail") }

func (s Suite) AssertErrorsAs(err error, target any) { // want AssertErrorsAs:`as\(0, 1\)`
	if !errors.As(err, target) {
		s.Fail()
	}
}

func (s Suite) AssertErrorsIs(err, target error) { // want AssertErrorsIs:`is\(0, 1\)`
	if !errors.Is(err, target) {
		s.Fail()
	}
}

func RequireErrorsAs(t *testing.T, err error, target any, format string, args ...any) { // want RequireErrorsAs:`as\(1, 2\)`
	t.Helper()

	if !errors.As(err, target) {
		t.Fatalf(format, args...)
	}
}

func RequireErrorsIs(t *testing.T, err, target error, format string, args ...any) { // want RequireErrorsIs:`is\(1, 2\)`
	t.Helper()

	if !errors.Is(err, target) {
		t.Fatalf(format, args...)
	}
}

func RequireErrorsAsType1[E error](t *testing.T, err error, format string, args ...any) { // want RequireErrorsAsType1:`astype\(1, 0\)`
	t.Helper()

	var target E
	if !errors.As(err, &target) {
		t.Fatalf(format, args...)
	}
}
