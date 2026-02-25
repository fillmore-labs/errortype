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

	"fillmore-labs.com/errortype/internal/detectwrappers/result"
	"fillmore-labs.com/errortype/internal/qualname"
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
	qualname.LocalFunc
	Type result.WrapperType
}

// _seeds are seed values that would otherwise not be recognized.
var _seeds = map[string][]ExplicitWrapper{
	"errors": {
		{qualname.LocalFunc{Name: "Is"}, result.WrapperIs},
		{qualname.LocalFunc{Name: "As"}, result.WrapperAs},
		{qualname.LocalFunc{Name: "AsType"}, result.WrapperAsType},
		{qualname.LocalFunc{Name: "Unwrap"}, result.WrapperUnwrap},
	},
	"fmt": {
		{qualname.LocalFunc{Name: "Errorf"}, result.WrapperErrorf},
	},
	"reflect": {
		{qualname.LocalFunc{Name: "TypeAssert"}, result.FuncAssert},
	},
	"golang.org/x/exp/errors": {
		{qualname.LocalFunc{Name: "Is"}, result.WrapperIs},
		{qualname.LocalFunc{Name: "As"}, result.WrapperAs},
		{qualname.LocalFunc{Name: "Unwrap"}, result.WrapperUnwrap},
	},
	"github.com/cockroachdb/errors/errutil": {
		{qualname.LocalFunc{Name: "As"}, result.WrapperAs},
	},
	"github.com/cockroachdb/errors/markers": {
		{qualname.LocalFunc{Name: "Is"}, result.WrapperIs},
	},
	"github.com/juju/errors": {
		{qualname.LocalFunc{Name: "AsType"}, result.WrapperAsType},
	},
	"github.com/stretchr/testify/assert": {
		{qualname.LocalFunc{Name: "IsType"}, result.FuncIsType},
		{qualname.LocalFunc{Name: "IsTypef"}, result.FuncIsType},
		{qualname.LocalFunc{Name: "IsNotType"}, result.FuncIsType},
		{qualname.LocalFunc{Name: "IsNotTypef"}, result.FuncIsType},
		{qualname.LocalFunc{Receiver: "Assertions", Name: "IsType"}, result.FuncIsType},
		{qualname.LocalFunc{Receiver: "Assertions", Name: "IsTypef"}, result.FuncIsType},
		{qualname.LocalFunc{Receiver: "Assertions", Name: "IsNotType"}, result.FuncIsType},
		{qualname.LocalFunc{Receiver: "Assertions", Name: "IsNotTypef"}, result.FuncIsType},
	},
	"github.com/stretchr/testify/require": {
		{qualname.LocalFunc{Name: "IsType"}, result.FuncIsType},
		{qualname.LocalFunc{Name: "IsTypef"}, result.FuncIsType},
		{qualname.LocalFunc{Name: "IsNotType"}, result.FuncIsType},
		{qualname.LocalFunc{Name: "IsNotTypef"}, result.FuncIsType},
		{qualname.LocalFunc{Receiver: "Assertions", Name: "IsType"}, result.FuncIsType},
		{qualname.LocalFunc{Receiver: "Assertions", Name: "IsTypef"}, result.FuncIsType},
		{qualname.LocalFunc{Receiver: "Assertions", Name: "IsNotType"}, result.FuncIsType},
		{qualname.LocalFunc{Receiver: "Assertions", Name: "IsNotTypef"}, result.FuncIsType},
	},
	"gotest.tools/v3/assert": {
		{qualname.LocalFunc{Name: "Equal"}, result.FuncEqual},
	},
	"gotest.tools/v3/assert/cmp": {
		{qualname.LocalFunc{Name: "Equal"}, result.FuncEqual},
	},
}
