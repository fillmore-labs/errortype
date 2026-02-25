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

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func wrap1(t assert.TestingT, err, target error) (assert.TestingT, error, error) {
	return t, err, target
}

func TestTestify(t *testing.T) {
	var err *my1Error

	_ = assert.ErrorIs(t, err, &my1Error{})         // want "is false or undefined"
	_ = assert.ErrorIsf(t, err, &my1Error{}, "")    // want "is false or undefined"
	_ = assert.NotErrorIs(t, err, &my1Error{})      // want "is false or undefined"
	_ = assert.NotErrorIsf(t, err, &my1Error{}, "") // want "is false or undefined"

	require.ErrorIs(t, err, &my1Error{})         // want "is false or undefined"
	require.ErrorIsf(t, err, &my1Error{}, "")    // want "is false or undefined"
	require.NotErrorIs(t, err, &my1Error{})      // want "is false or undefined"
	require.NotErrorIsf(t, err, &my1Error{}, "") // want "is false or undefined"

	_ = assert.ErrorIs(wrap1(t, err, &my1Error{}))

	_ = assert.Equal(t, err, &my1Error{})

	_ = assert.Error(t, &my1Error{})

	var s suite.Suite

	_ = s.ErrorIs(err, &my1Error{})         // want "is false or undefined"
	_ = s.ErrorIsf(err, &my1Error{}, "")    // want "is false or undefined"
	_ = s.NotErrorIs(err, &my1Error{})      // want "is false or undefined"
	_ = s.NotErrorIsf(err, &my1Error{}, "") // want "is false or undefined"

	r := s.Require()

	r.ErrorIs(err, &my1Error{})         // want "is false or undefined"
	r.ErrorIsf(err, &my1Error{}, "")    // want "is false or undefined"
	r.NotErrorIs(err, &my1Error{})      // want "is false or undefined"
	r.NotErrorIsf(err, &my1Error{}, "") // want "is false or undefined"

	_ = (*assert.Assertions).ErrorIs(s.Assertions, err, &my1Error{}) // want "is false or undefined"
	(*require.Assertions).ErrorIs(r, err, &my1Error{})               // want "is false or undefined"

	_ = s.Equal(err, &my1Error{})
	r.Equal(err, &my1Error{})
}
