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
	var err *myError1

	assert.ErrorIs(t, err, &myError1{})         // want "is false or undefined"
	assert.ErrorIsf(t, err, &myError1{}, "")    // want "is false or undefined"
	assert.NotErrorIs(t, err, &myError1{})      // want "is false or undefined"
	assert.NotErrorIsf(t, err, &myError1{}, "") // want "is false or undefined"

	require.ErrorIs(t, err, &myError1{})         // want "is false or undefined"
	require.ErrorIsf(t, err, &myError1{}, "")    // want "is false or undefined"
	require.NotErrorIs(t, err, &myError1{})      // want "is false or undefined"
	require.NotErrorIsf(t, err, &myError1{}, "") // want "is false or undefined"

	assert.ErrorIs(wrap1(t, err, &myError1{}))

	assert.Equal(t, err, &myError1{})

	assert.Error(t, &myError1{})

	var s suite.Suite
	r := s.Require()

	s.ErrorIs(err, &myError1{})         // want "is false or undefined"
	s.ErrorIsf(err, &myError1{}, "")    // want "is false or undefined"
	s.NotErrorIs(err, &myError1{})      // want "is false or undefined"
	s.NotErrorIsf(err, &myError1{}, "") // want "is false or undefined"

	r.ErrorIs(err, &myError1{})         // want "is false or undefined"
	r.ErrorIsf(err, &myError1{}, "")    // want "is false or undefined"
	r.NotErrorIs(err, &myError1{})      // want "is false or undefined"
	r.NotErrorIsf(err, &myError1{}, "") // want "is false or undefined"

	(*assert.Assertions).ErrorIs(s.Assertions, err, &myError1{}) // want "is false or undefined"
	(*require.Assertions).ErrorIs(r, err, &myError1{})           // want "is false or undefined"

	s.Equal(err, &myError1{})
	r.Equal(err, &myError1{})
}
