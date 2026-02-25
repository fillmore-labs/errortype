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
	"go/types"

	"fillmore-labs.com/errortype/detect/result"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// LookupSeeds retrieves known error wrapper functions for a given package using predefined seed values.
func LookupSeeds(pkg *types.Package) result.ErrorFuncs {
	pkgFuncs, ok := _seeds[pkg.Path()]
	if !ok {
		return nil
	}

	return LookupWrappers(pkg, pkgFuncs, false)
}

// ExplicitWrapper defines a function to be recognized as a specific type of error wrapper.
type ExplicitWrapper struct {
	typeutil.LocalFuncName
	Type result.WrapperType
}

// _seeds are seed values that would otherwise not be recognized.
var _seeds = map[string][]ExplicitWrapper{
	"errors": {
		{typeutil.LocalFuncName{Name: "Is"}, result.WrapperIs},
		{typeutil.LocalFuncName{Name: "As"}, result.WrapperAs},
		{typeutil.LocalFuncName{Name: "AsType"}, result.WrapperAsType},
	},
	"fmt": {
		{typeutil.LocalFuncName{Name: "Errorf"}, result.WrapperErrorf},
	},
	"reflect": {
		{typeutil.LocalFuncName{Name: "TypeAssert"}, result.FuncAssert},
	},
	"golang.org/x/exp/errors": {
		{typeutil.LocalFuncName{Name: "Is"}, result.WrapperIs},
		{typeutil.LocalFuncName{Name: "As"}, result.WrapperAs},
	},
	"github.com/cockroachdb/errors/errutil": {
		{typeutil.LocalFuncName{Name: "As"}, result.WrapperAs},
	},
	"github.com/cockroachdb/errors/markers": {
		{typeutil.LocalFuncName{Name: "Is"}, result.WrapperIs},
	},
	"github.com/juju/errors": {
		{typeutil.LocalFuncName{Name: "AsType"}, result.WrapperAsType},
	},
	"github.com/stretchr/testify/assert": {
		{typeutil.LocalFuncName{Name: "IsType"}, result.FuncIsType},
		{typeutil.LocalFuncName{Name: "IsTypef"}, result.FuncIsType},
		{typeutil.LocalFuncName{Name: "IsNotType"}, result.FuncIsType},
		{typeutil.LocalFuncName{Name: "IsNotTypef"}, result.FuncIsType},
		{typeutil.LocalFuncName{Receiver: "Assertions", Name: "IsType"}, result.FuncIsType},
		{typeutil.LocalFuncName{Receiver: "Assertions", Name: "IsTypef"}, result.FuncIsType},
		{typeutil.LocalFuncName{Receiver: "Assertions", Name: "IsNotType"}, result.FuncIsType},
		{typeutil.LocalFuncName{Receiver: "Assertions", Name: "IsNotTypef"}, result.FuncIsType},
	},
	"github.com/stretchr/testify/require": {
		{typeutil.LocalFuncName{Name: "IsType"}, result.FuncIsType},
		{typeutil.LocalFuncName{Name: "IsTypef"}, result.FuncIsType},
		{typeutil.LocalFuncName{Name: "IsNotType"}, result.FuncIsType},
		{typeutil.LocalFuncName{Name: "IsNotTypef"}, result.FuncIsType},
		{typeutil.LocalFuncName{Receiver: "Assertions", Name: "IsType"}, result.FuncIsType},
		{typeutil.LocalFuncName{Receiver: "Assertions", Name: "IsTypef"}, result.FuncIsType},
		{typeutil.LocalFuncName{Receiver: "Assertions", Name: "IsNotType"}, result.FuncIsType},
		{typeutil.LocalFuncName{Receiver: "Assertions", Name: "IsNotTypef"}, result.FuncIsType},
	},
	"gotest.tools/v3/assert": {
		{typeutil.LocalFuncName{Name: "Equal"}, result.FuncEqual},
	},
	"gotest.tools/v3/assert/cmp": {
		{typeutil.LocalFuncName{Name: "Equal"}, result.FuncEqual},
	},
}
