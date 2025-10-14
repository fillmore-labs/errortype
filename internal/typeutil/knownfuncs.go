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

// KnownFuncs maps function metadata (FuncName) to a packed integer containing
// information about the function's behavior (e.g., if it's like errors.Is or errors.As).
var KnownFuncs = map[FuncName]FuncInfo{
	// errors.Is-like functions
	{Path: "errors", Name: "Is"}:                                                                          isFuncType0Result,
	{Path: "golang.org/x/exp/errors", Name: "Is"}:                                                         isFuncType0Result,
	{Path: "golang.org/x/xerrors", Name: "Is"}:                                                            isFuncType0Result,
	{Path: "github.com/pkg/errors", Name: "Is"}:                                                           isFuncType0Result,
	{Path: "github.com/friendsofgo/errors", Name: "Is"}:                                                   isFuncType0Result,
	{Path: "github.com/go-errors/errors", Name: "Is"}:                                                     isFuncType0Result,
	{Path: "github.com/go-faster/errors", Name: "Is"}:                                                     isFuncType0Result,
	{Path: "github.com/cockroachdb/errors", Name: "Is"}:                                                   isFuncType0Result,
	{Path: "github.com/cockroachdb/errors/markers", Name: "Is"}:                                           isFuncType0Result,
	{Path: "github.com/juju/errors", Name: "Is"}:                                                          isFuncType0Result,
	{Path: "gotest.tools/v3/assert", Name: "Equal"}:                                                       equFuncType1Ignore,
	{Path: "gotest.tools/v3/assert", Name: "ErrorIs"}:                                                     isFuncType1Ignore,
	{Path: "gotest.tools/v3/assert/cmp", Name: "Equal"}:                                                   equFuncType0Ignore,
	{Path: "github.com/stretchr/testify/assert", Name: "ErrorIs"}:                                         isFuncType1Ignore,
	{Path: "github.com/stretchr/testify/assert", Name: "ErrorIsf"}:                                        isFuncType1Ignore,
	{Path: "github.com/stretchr/testify/assert", Name: "NotErrorIs"}:                                      isFuncType1Ignore,
	{Path: "github.com/stretchr/testify/assert", Name: "NotErrorIsf"}:                                     isFuncType1Ignore,
	{Path: "github.com/stretchr/testify/require", Name: "ErrorIs"}:                                        isFuncType1Ignore,
	{Path: "github.com/stretchr/testify/require", Name: "ErrorIsf"}:                                       isFuncType1Ignore,
	{Path: "github.com/stretchr/testify/require", Name: "NotErrorIs"}:                                     isFuncType1Ignore,
	{Path: "github.com/stretchr/testify/require", Name: "NotErrorIsf"}:                                    isFuncType1Ignore,
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "ErrorIs", Ptr: true}:      isFuncType0Ignore,
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "ErrorIsf", Ptr: true}:     isFuncType0Ignore,
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "NotErrorIs", Ptr: true}:   isFuncType0Ignore,
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "NotErrorIsf", Ptr: true}:  isFuncType0Ignore,
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "ErrorIs", Ptr: true}:     isFuncType0Ignore,
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "ErrorIsf", Ptr: true}:    isFuncType0Ignore,
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "NotErrorIs", Ptr: true}:  isFuncType0Ignore,
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "NotErrorIsf", Ptr: true}: isFuncType0Ignore,

	// errors.As-like functions
	{Path: "errors", Name: "As"}:                                                                          asFuncType0Result,
	{Path: "errors", Name: "AsType"}:                                                                      assertFuncWithType,
	{Path: "reflect", Name: "TypeAssert"}:                                                                 assertFuncWithType,
	{Path: "fillmore-labs.com/exp/errors", Name: "Has"}:                                                   assertFuncWithType,
	{Path: "fillmore-labs.com/exp/errors", Name: "HasError"}:                                              assertFuncWithType,
	{Path: "fillmore-labs.com/exp/errors", Name: "As"}:                                                    asFuncType0WithType,
	{Path: "fillmore-labs.com/exp/errors", Name: "AsError"}:                                               asFuncType0WithType,
	{Path: "golang.org/x/exp/errors", Name: "As"}:                                                         asFuncType0Result,
	{Path: "golang.org/x/xerrors", Name: "As"}:                                                            asFuncType0Result,
	{Path: "github.com/pkg/errors", Name: "As"}:                                                           asFuncType0Result,
	{Path: "github.com/friendsofgo/errors", Name: "As"}:                                                   asFuncType0Result,
	{Path: "github.com/go-errors/errors", Name: "As"}:                                                     asFuncType0Result,
	{Path: "github.com/go-faster/errors", Name: "As"}:                                                     asFuncType0Result,
	{Path: "github.com/cockroachdb/errors", Name: "As"}:                                                   asFuncType0Result,
	{Path: "github.com/cockroachdb/errors/errutil", Name: "As"}:                                           asFuncType0Result,
	{Path: "github.com/juju/errors", Name: "As"}:                                                          asFuncType0Result,
	{Path: "github.com/juju/errors", Name: "AsType"}:                                                      assertFuncWithType,
	{Path: "github.com/juju/errors", Name: "HasType"}:                                                     assertFuncWithType,
	{Path: "github.com/stretchr/testify/assert", Name: "ErrorAs"}:                                         asFuncType1Ignore,
	{Path: "github.com/stretchr/testify/assert", Name: "ErrorAsf"}:                                        asFuncType1Ignore,
	{Path: "github.com/stretchr/testify/assert", Name: "NotErrorAs"}:                                      asFuncType1Ignore,
	{Path: "github.com/stretchr/testify/assert", Name: "NotErrorAsf"}:                                     asFuncType1Ignore,
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "ErrorAs", Ptr: true}:      asFuncType0Ignore,
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "ErrorAsf", Ptr: true}:     asFuncType0Ignore,
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "NotErrorAs", Ptr: true}:   asFuncType0Ignore,
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "NotErrorAsf", Ptr: true}:  asFuncType0Ignore,
	{Path: "github.com/stretchr/testify/require", Name: "ErrorAs"}:                                        asFuncType1Ignore,
	{Path: "github.com/stretchr/testify/require", Name: "ErrorAsf"}:                                       asFuncType1Ignore,
	{Path: "github.com/stretchr/testify/require", Name: "NotErrorAs"}:                                     asFuncType1Ignore,
	{Path: "github.com/stretchr/testify/require", Name: "NotErrorAsf"}:                                    asFuncType1Ignore,
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "ErrorAs", Ptr: true}:     asFuncType0Ignore,
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "ErrorAsf", Ptr: true}:    asFuncType0Ignore,
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "NotErrorAs", Ptr: true}:  asFuncType0Ignore,
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "NotErrorAsf", Ptr: true}: asFuncType0Ignore,

	// assert.IsType-like functions
	{Path: "github.com/stretchr/testify/assert", Name: "IsType"}:                                         typFuncType1Ignore,
	{Path: "github.com/stretchr/testify/assert", Name: "IsTypef"}:                                        typFuncType1Ignore,
	{Path: "github.com/stretchr/testify/assert", Name: "IsNotType"}:                                      typFuncType1Ignore,
	{Path: "github.com/stretchr/testify/assert", Name: "IsNotTypef"}:                                     typFuncType1Ignore,
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "IsType", Ptr: true}:      typFuncType0Ignore,
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "IsTypef", Ptr: true}:     typFuncType0Ignore,
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "IsNotType", Ptr: true}:   typFuncType0Ignore,
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "IsNotTypef", Ptr: true}:  typFuncType0Ignore,
	{Path: "github.com/stretchr/testify/require", Name: "IsType"}:                                        typFuncType1Ignore,
	{Path: "github.com/stretchr/testify/require", Name: "IsTypef"}:                                       typFuncType1Ignore,
	{Path: "github.com/stretchr/testify/require", Name: "IsNotType"}:                                     typFuncType1Ignore,
	{Path: "github.com/stretchr/testify/require", Name: "IsNotTypef"}:                                    typFuncType1Ignore,
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "IsType", Ptr: true}:     typFuncType0Ignore,
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "IsTypef", Ptr: true}:    typFuncType0Ignore,
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "IsNotType", Ptr: true}:  typFuncType0Ignore,
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "IsNotTypef", Ptr: true}: typFuncType0Ignore,
}
