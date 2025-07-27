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

package a

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ValueError struct {
	Msg string
}

func (v ValueError) Error() string {
	return v.Msg
}

var _ error = ValueError{}

func TestTestify(t *testing.T) {
	var (
		err error
		pve *ValueError
	)

	_ = assert.ErrorAs(t, err, &pve)         // want " \\(et:err\\)$"
	_ = assert.ErrorAsf(t, err, &pve, "")    // want " \\(et:err\\)$"
	_ = assert.NotErrorAs(t, err, &pve)      // want " \\(et:err\\)$"
	_ = assert.NotErrorAsf(t, err, &pve, "") // want " \\(et:err\\)$"

	a := assert.New(t)

	_ = a.ErrorAs(err, &pve)         // want " \\(et:err\\)$"
	_ = a.ErrorAs(err, &pve, "")     // want " \\(et:err\\)$"
	_ = a.ErrorAsf(err, &pve, "")    // want " \\(et:err\\)$"
	_ = a.NotErrorAs(err, &pve)      // want " \\(et:err\\)$"
	_ = a.NotErrorAsf(err, &pve, "") // want " \\(et:err\\)$"

	_ = (*assert.Assertions).ErrorAs(a, err, &pve) // want " \\(et:err\\)$"

	require.ErrorAs(t, err, &pve)         // want " \\(et:err\\)$"
	require.ErrorAs(t, err, &pve, "")     // want " \\(et:err\\)$"
	require.ErrorAsf(t, err, &pve, "")    // want " \\(et:err\\)$"
	require.NotErrorAs(t, err, &pve)      // want " \\(et:err\\)$"
	require.NotErrorAsf(t, err, &pve, "") // want " \\(et:err\\)$"

	r := require.New(t)

	r.ErrorAs(err, &pve)         // want " \\(et:err\\)$"
	r.ErrorAs(err, &pve, "")     // want " \\(et:err\\)$"
	r.ErrorAsf(err, &pve, "")    // want " \\(et:err\\)$"
	r.NotErrorAs(err, &pve)      // want " \\(et:err\\)$"
	r.NotErrorAsf(err, &pve, "") // want " \\(et:err\\)$"

	(*require.Assertions).ErrorAs(r, err, &pve) // want " \\(et:err\\)$"
}

type MySuite struct {
	suite.Suite
}

func TestMyTestSuite(t *testing.T) {
	suite.Run(t, &MySuite{})
}

func (s *MySuite) TestMySuite() {
	var (
		err error
		pve *ValueError
	)

	s.ErrorAs(err, &pve)                   // want " \\(et:err\\)$"
	s.ErrorAsf(err, &pve, "")              // want " \\(et:err\\)$"
	s.NotErrorAs(err, &pve)                // want " \\(et:err\\)$"
	s.NotErrorAsf(err, &pve, "")           // want " \\(et:err\\)$"
	s.Require().ErrorAs(err, &pve)         // want " \\(et:err\\)$"
	s.Require().ErrorAsf(err, &pve, "")    // want " \\(et:err\\)$"
	s.Require().NotErrorAs(err, &pve)      // want " \\(et:err\\)$"
	s.Require().NotErrorAsf(err, &pve, "") // want " \\(et:err\\)$"
}
