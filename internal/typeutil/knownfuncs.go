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
	{Path: "errors", Name: "Is"}:                                                                          IsFuncType0,
	{Path: "golang.org/x/exp/errors", Name: "Is"}:                                                         IsFuncType0,
	{Path: "golang.org/x/xerrors", Name: "Is"}:                                                            IsFuncType0,
	{Path: "github.com/pkg/errors", Name: "Is"}:                                                           IsFuncType0,
	{Path: "github.com/go-errors/errors", Name: "Is"}:                                                     IsFuncType0,
	{Path: "github.com/go-faster/errors", Name: "Is"}:                                                     IsFuncType0,
	{Path: "github.com/cockroachdb/errors", Name: "Is"}:                                                   IsFuncType0,
	{Path: "github.com/juju/errors", Name: "Is"}:                                                          IsFuncType0,
	{Path: "gotest.tools/v3/assert", Name: "Equal"}:                                                       EquFuncType1,
	{Path: "gotest.tools/v3/assert", Name: "ErrorIs"}:                                                     IsFuncType1,
	{Path: "github.com/stretchr/testify/assert", Name: "ErrorIs"}:                                         IsFuncType1,
	{Path: "github.com/stretchr/testify/assert", Name: "ErrorIsf"}:                                        IsFuncType1,
	{Path: "github.com/stretchr/testify/assert", Name: "NotErrorIs"}:                                      IsFuncType1,
	{Path: "github.com/stretchr/testify/assert", Name: "NotErrorIsf"}:                                     IsFuncType1,
	{Path: "github.com/stretchr/testify/require", Name: "ErrorIs"}:                                        IsFuncType1,
	{Path: "github.com/stretchr/testify/require", Name: "ErrorIsf"}:                                       IsFuncType1,
	{Path: "github.com/stretchr/testify/require", Name: "NotErrorIs"}:                                     IsFuncType1,
	{Path: "github.com/stretchr/testify/require", Name: "NotErrorIsf"}:                                    IsFuncType1,
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "ErrorIs", Ptr: true}:      IsFuncType0,
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "ErrorIsf", Ptr: true}:     IsFuncType0,
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "NotErrorIs", Ptr: true}:   IsFuncType0,
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "NotErrorIsf", Ptr: true}:  IsFuncType0,
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "ErrorIs", Ptr: true}:     IsFuncType0,
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "ErrorIsf", Ptr: true}:    IsFuncType0,
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "NotErrorIs", Ptr: true}:  IsFuncType0,
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "NotErrorIsf", Ptr: true}: IsFuncType0,

	// errors.As-like functions
	{Path: "errors", Name: "As"}:                                                                          AsFuncType0,
	{Path: "errors", Name: "AsType"}:                                                                      AssertFuncWithType,
	{Path: "reflect", Name: "TypeAssert"}:                                                                 AssertFuncWithType,
	{Path: "fillmore-labs.com/exp/errors", Name: "Has"}:                                                   AssertFuncWithType,
	{Path: "fillmore-labs.com/exp/errors", Name: "HasError"}:                                              AssertFuncWithType,
	{Path: "fillmore-labs.com/exp/errors", Name: "As"}:                                                    AsFuncType0WithType,
	{Path: "fillmore-labs.com/exp/errors", Name: "AsError"}:                                               AsFuncType0WithType,
	{Path: "golang.org/x/exp/errors", Name: "As"}:                                                         AsFuncType0,
	{Path: "golang.org/x/xerrors", Name: "As"}:                                                            AsFuncType0,
	{Path: "github.com/pkg/errors", Name: "As"}:                                                           AsFuncType0,
	{Path: "github.com/go-errors/errors", Name: "As"}:                                                     AsFuncType0,
	{Path: "github.com/go-faster/errors", Name: "As"}:                                                     AsFuncType0,
	{Path: "github.com/cockroachdb/errors", Name: "As"}:                                                   AsFuncType0,
	{Path: "github.com/cockroachdb/errors/errutil", Name: "As"}:                                           AsFuncType0,
	{Path: "github.com/juju/errors", Name: "As"}:                                                          AsFuncType0,
	{Path: "github.com/juju/errors", Name: "AsType"}:                                                      AssertFuncWithType,
	{Path: "github.com/juju/errors", Name: "HasType"}:                                                     AssertFuncWithType,
	{Path: "github.com/stretchr/testify/assert", Name: "ErrorAs"}:                                         AsFuncType1,
	{Path: "github.com/stretchr/testify/assert", Name: "ErrorAsf"}:                                        AsFuncType1,
	{Path: "github.com/stretchr/testify/assert", Name: "NotErrorAs"}:                                      AsFuncType1,
	{Path: "github.com/stretchr/testify/assert", Name: "NotErrorAsf"}:                                     AsFuncType1,
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "ErrorAs", Ptr: true}:      AsFuncType0,
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "ErrorAsf", Ptr: true}:     AsFuncType0,
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "NotErrorAs", Ptr: true}:   AsFuncType0,
	{Path: "github.com/stretchr/testify/assert", Receiver: "Assertions", Name: "NotErrorAsf", Ptr: true}:  AsFuncType0,
	{Path: "github.com/stretchr/testify/require", Name: "ErrorAs"}:                                        AsFuncType1,
	{Path: "github.com/stretchr/testify/require", Name: "ErrorAsf"}:                                       AsFuncType1,
	{Path: "github.com/stretchr/testify/require", Name: "NotErrorAs"}:                                     AsFuncType1,
	{Path: "github.com/stretchr/testify/require", Name: "NotErrorAsf"}:                                    AsFuncType1,
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "ErrorAs", Ptr: true}:     AsFuncType0,
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "ErrorAsf", Ptr: true}:    AsFuncType0,
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "NotErrorAs", Ptr: true}:  AsFuncType0,
	{Path: "github.com/stretchr/testify/require", Receiver: "Assertions", Name: "NotErrorAsf", Ptr: true}: AsFuncType0,
}
