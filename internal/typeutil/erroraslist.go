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

package typeutil

import (
	"go/ast"
	"go/types"
)

// IsErrorAs analyzes a function call to determine if it matches patterns like errors.As and identifies the target argument.
// It returns the resolved function and the index of its target argument, or nil, -1 if the function is not of interest.
func IsErrorAs(info *types.Info, n *ast.CallExpr) (fun *types.Func, targetType ast.Expr, targetArgIndex int) {
	fun, typeParams, methodExpr, ok := FuncOf(info, n.Fun)
	if !ok {
		return nil, nil, -1 // Could not resolve function, might be a func variable.
	}

	// Check if the function is one we analyze (e.g., errors.As).
	// errorsAs maps a function name to the index of its "target" argument.
	funcName := FuncNameOf(fun)

	target, ok := errorsAs[funcName]
	if !ok {
		return nil, nil, -1 // Not a function we are interested in.
	}

	if target.typeParam >= 0 {
		if len(typeParams) <= target.typeParam {
			return nil, nil, -1 // Not enough type parameters
		}
		typ := typeParams[target.typeParam]

		return fun, typ, -1
	}

	targetArgIndex = target.targetArgIndex

	if methodExpr {
		// For method expression calls ("(*assert.Assertions).ErrorsAs(a, ...)"),
		// the receiver `a` is the first argument. The argument indices in `errorsAs`
		// are for the function form, so we increment the index to correctly locate
		// the target argument in the method call expression.
		targetArgIndex++
	}

	return fun, nil, targetArgIndex
}

// errorsAs maps functions that behave like errors.As to the argument index
// of their "target" parameter. This allows the analyzer to identify which
// argument in a call to these functions should be checked for correct
// pointer-vs-value usage.
var errorsAs = map[FuncName]struct{ targetArgIndex, typeParam int }{
	{Path: "errors", Name: "As"}:                                                                          {1, -1},
	{Path: "reflect", Name: "TypeAssert"}:                                                                 {-1, 0},
	{Path: "golang.org/x/exp/errors", Name: "As"}:                                                         {1, -1},
	{Path: "golang.org/x/xerrors", Name: "As"}:                                                            {1, -1},
	{Path: "github.com/pkg/errors", Name: "As"}:                                                           {1, -1},
	{Path: "github.com/go-errors/errors", Name: "As"}:                                                     {1, -1},
	{Path: "github.com/cockroachdb/errors", Name: "As"}:                                                   {1, -1},
	{Path: "github.com/cockroachdb/errors/errutil", Name: "As"}:                                           {1, -1},
	{Path: "github.com/juju/errors", Name: "As"}:                                                          {1, -1},
	{Path: "github.com/juju/errors", Name: "AsType"}:                                                      {-1, 0},
	{Path: "github.com/juju/errors", Name: "HasType"}:                                                     {-1, 0},
	{Path: "github.com/stretchr/testify/assert", Name: "ErrorAs"}:                                         {2, -1},
	{Path: "github.com/stretchr/testify/assert", Name: "ErrorAsf"}:                                        {2, -1},
	{Path: "github.com/stretchr/testify/assert", Name: "NotErrorAs"}:                                      {2, -1},
	{Path: "github.com/stretchr/testify/assert", Name: "NotErrorAsf"}:                                     {2, -1},
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "ErrorAs", Ptr: true}:      {1, -1},
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "ErrorAsf", Ptr: true}:     {1, -1},
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "NotErrorAs", Ptr: true}:   {1, -1},
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "NotErrorAsf", Ptr: true}:  {1, -1},
	{Path: "github.com/stretchr/testify/require", Name: "ErrorAs"}:                                        {2, -1},
	{Path: "github.com/stretchr/testify/require", Name: "ErrorAsf"}:                                       {2, -1},
	{Path: "github.com/stretchr/testify/require", Name: "NotErrorAs"}:                                     {2, -1},
	{Path: "github.com/stretchr/testify/require", Name: "NotErrorAsf"}:                                    {2, -1},
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "ErrorAs", Ptr: true}:     {1, -1},
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "ErrorAsf", Ptr: true}:    {1, -1},
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "NotErrorAs", Ptr: true}:  {1, -1},
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "NotErrorAsf", Ptr: true}: {1, -1},
}
