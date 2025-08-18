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

// FuncKind distinguishes between different kinds of special functions.
type FuncKind int

const (
	// indicates the function is unconfigured.
	_ FuncKind = iota

	// KindIs indicates a function that behaves like errors.Is.
	KindIs

	// KindAs indicates a function that behaves like errors.As.
	KindAs

	// KindEqu indicates a function that behaves like assert.Equal.
	KindEqu
)

// FuncType defines a type used to classify specific function behaviors or roles in analysis and comparisons.
type FuncType int8

const (
	// the unclassified state for a function type, this should not happen.
	_ FuncType = iota

	// IsFunc0 represents a function type that performs error comparison with no additional context or parameters.
	IsFunc0

	// IsFunc1 represents a function type that performs comparison with one additional context parameter, typically for error assertion functions.
	IsFunc1
)

const (
	// the unclassified state for a function type, this should not happen.
	_ int8 = iota

	// AsFunc0 represents a function type that performs error assertion with no additional context or parameters.
	AsFunc0

	// AsFunc1 represents a function type that performs error assertion with one additional context parameter.
	AsFunc1

	// AssertFunc represents a function type that performs error assertion without a target parameter.
	AssertFunc int8 = -1
)

const (
	// AsTypeParam0 represents a function type that performs error assertion with a type parameter.
	AsTypeParam0 int8 = 0

	// AsTypeParamNone represents a function type that performs error assertion without a type parameter.
	AsTypeParamNone int8 = -1
)

// FuncInfo holds packed information about special functions like errors.Is or errors.As.
// It is designed to be stored as a single integer value in a map for efficiency.
//
// The bit layout is as follows:
//   - Bits 0-1: FuncKind (2 bits, for KindIs, KindAs, KindCmp)
//
// If KindIs:
//   - Bits 2-3: FuncType (2 bits)
//
// If KindAs:
//   - Bits 2-3: TargetArgIndex + 1 (4 bits, allows index from -1 to 2)
//   - Bits 4-5: TypeParam + 1 (2 bits, allows index from -1 to 2)
type FuncInfo int8

const (
	kindMask = 0b11

	isTypeShift = 2
	isTypeMask  = 0b11

	asTargetArgShift = 2
	asTargetArgMask  = 0b11
	asTypeParamShift = 4
	asTypeParamMask  = 0b11

	asIndexOffset = 1
)

// Pre-defined FuncInfo constants for common function configurations.
const (
	// Is-like functions.
	IsFuncType0 = FuncInfo(KindIs) | (FuncInfo(IsFunc0) << isTypeShift)
	IsFuncType1 = FuncInfo(KindIs) | (FuncInfo(IsFunc1) << isTypeShift)

	// Equal-like functions.
	EquFuncType1 = FuncInfo(KindEqu) | (FuncInfo(IsFunc1) << isTypeShift)

	// As-like functions.
	AsFuncType0         = FuncInfo(KindAs) | (FuncInfo(AsFunc0+asIndexOffset) << asTargetArgShift) | (FuncInfo(AsTypeParamNone+asIndexOffset) << asTypeParamShift)
	AsFuncType1         = FuncInfo(KindAs) | (FuncInfo(AsFunc1+asIndexOffset) << asTargetArgShift) | (FuncInfo(AsTypeParamNone+asIndexOffset) << asTypeParamShift)
	AsFuncType0WithType = FuncInfo(KindAs) | (FuncInfo(AsFunc0+asIndexOffset) << asTargetArgShift) | (FuncInfo(AsTypeParam0+asIndexOffset) << asTypeParamShift)

	// Assert-like functions.
	AssertFuncWithType = FuncInfo(KindAs) | (FuncInfo(AssertFunc+asIndexOffset) << asTargetArgShift) | (FuncInfo(AsTypeParam0+asIndexOffset) << asTypeParamShift)
)

// Kind returns the kind of the function (e.g., KindIs, KindAs).
func (i FuncInfo) Kind() FuncKind {
	return FuncKind(i & kindMask)
}

// IsType returns the FuncType for an errors.Is-like function.
// It should only be called if Kind() is KindIs.
func (i FuncInfo) IsType() FuncType {
	return FuncType((i >> isTypeShift) & isTypeMask)
}

// AsTarget returns the TargetArgIndex and TypeParam for an errors.As-like function.
// It should only be called if Kind() is KindAs.
func (i FuncInfo) AsTarget() (targetArgIndex, typeParam int) {
	targetArgIndex = int((i>>asTargetArgShift)&asTargetArgMask) - asIndexOffset
	typeParam = int((i>>asTypeParamShift)&asTypeParamMask) - asIndexOffset

	return targetArgIndex, typeParam
}
