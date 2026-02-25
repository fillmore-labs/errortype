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

package rules

import (
	"go/types"
	"strings"

	"fillmore-labs.com/errortype/internal/typeutil"
)

const (
	// VarPrefix is the prefix of unexported error variables.
	VarPrefix = "err"
	// VarPrefixExported is the prefix of exported error variables.
	VarPrefixExported = "Err"
	// TypeSuffix is the suffix of error types.
	TypeSuffix = "Error"
	// TypeSuffixMultiple is accepted as suffix of error types with an "Unwrap() []error" method, but not suggested.
	TypeSuffixMultiple = "Errors"
)

func conformantVar(def types.Object) bool {
	name := def.Name()
	if def.Exported() {
		return strings.HasPrefix(name, VarPrefixExported)
	}

	return strings.HasPrefix(name, VarPrefix)
}

func conformantType(tn *types.TypeName) bool {
	name := tn.Name()

	return strings.HasSuffix(name, TypeSuffix) ||
		strings.HasSuffix(name, TypeSuffixMultiple) &&
			typeutil.UnwrapMultipleMethod.HasMethod(tn.Type(), true)
}
