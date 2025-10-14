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

package knownfuncs

// FuncKind distinguishes between different kinds of special functions.
type FuncKind int8

const (
	// Indicates the function is unconfigured.
	_ FuncKind = iota

	// KindIs indicates a function that behaves like errors.Is.
	KindIs

	// KindAs indicates a function that behaves like errors.As.
	KindAs

	// KindEqu indicates a function that behaves like assert.Equal.
	KindEqu

	// KindType indicates a function that behaves like assert.IsType.
	KindType
)

// FuncType defines a type used to classify specific function behaviors or roles in analysis and comparisons.
type FuncType = int8

const (
	// The unclassified state for a function type, this should not happen.
	_ FuncType = iota

	// IsFunc0 represents a function type that performs error comparison with no additional context or parameters.
	IsFunc0

	// IsFunc1 represents a function type that performs comparison with one additional context parameter, typically for error assertion functions.
	IsFunc1
)

const (
	// The unclassified state for a function type, this should not happen.
	_ FuncType = iota

	// AsFunc0 represents a function type that performs error assertion with no additional context or parameters.
	AsFunc0

	// AsFunc1 represents a function type that performs error assertion with one additional context parameter.
	AsFunc1

	// AssertFunc represents a function type that performs error assertion without a target parameter.
	AssertFunc FuncType = -1
)

const (
	// AsTypeParam0 represents a function type that performs error assertion with a type parameter.
	AsTypeParam0 int8 = 0

	// AsTypeParamNone represents a function type that performs error assertion without a type parameter.
	AsTypeParamNone int8 = -1
)

// EvalType represents the need for the result of a function to be evaluated.
type EvalType = int8

const (
	_ EvalType = iota

	// MustEval represents a function type that results must be evaluated.
	MustEval

	// ShouldEval represents a function type that results should be evaluated.
	ShouldEval
)

// FuncInfo holds packed information about special functions like errors.Is or errors.As.
type FuncInfo struct {
	kind      FuncKind
	targetArg FuncType
	typeParam int8
	eval      EvalType
}

// Kind returns the kind of the function (e.g., KindIs, KindAs).
func (i FuncInfo) Kind() FuncKind {
	return i.kind
}

// IsType returns the FuncType for an `errors.Is`-like function.
// It should only be called if Kind() is KindIs.
func (i FuncInfo) IsType() FuncType {
	return i.targetArg
}

// AsTarget returns the TargetArgIndex and TypeParam for an `errors.As`-like function.
// It should only be called if Kind() is KindAs.
func (i FuncInfo) AsTarget() (targetArgIndex, typeParam int) {
	return int(i.targetArg), int(i.typeParam)
}

// EvalType is true when the result of the function should not be ignored.
func (i FuncInfo) EvalType() EvalType {
	return i.eval
}

// Pre-defined FuncInfo constants for common function configurations.
var (
	// Is-like functions.
	isFuncType0Ignore = FuncInfo{kind: KindIs, targetArg: IsFunc0}
	isFuncType0Result = FuncInfo{kind: KindIs, targetArg: IsFunc0, eval: MustEval}
	isFuncType1Ignore = FuncInfo{kind: KindIs, targetArg: IsFunc1}

	// Equal-like functions.
	equFuncType0Ignore = FuncInfo{kind: KindEqu, targetArg: IsFunc0}
	equFuncType1Ignore = FuncInfo{kind: KindEqu, targetArg: IsFunc1}

	// As-like functions.
	asFuncType0Ignore   = FuncInfo{kind: KindAs, targetArg: AsFunc0, typeParam: AsTypeParamNone}
	asFuncType0Result   = FuncInfo{kind: KindAs, targetArg: AsFunc0, typeParam: AsTypeParamNone, eval: ShouldEval}
	asFuncType1Ignore   = FuncInfo{kind: KindAs, targetArg: AsFunc1, typeParam: AsTypeParamNone}
	asFuncType0WithType = FuncInfo{kind: KindAs, targetArg: AsFunc0, typeParam: AsTypeParam0, eval: ShouldEval}

	// Assert-like functions without a target parameter.
	assertFuncWithType = FuncInfo{kind: KindAs, targetArg: AssertFunc, typeParam: AsTypeParam0, eval: MustEval}

	// IsType-like functions.
	typFuncType0Ignore = FuncInfo{kind: KindType, targetArg: IsFunc0}
	typFuncType1Ignore = FuncInfo{kind: KindType, targetArg: IsFunc1}
)
