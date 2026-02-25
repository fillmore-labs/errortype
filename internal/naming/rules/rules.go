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
	"unicode"
	"unicode/utf8"

	"fillmore-labs.com/errortype/internal/errorlike"
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

	prefix := VarPrefix
	if def.Exported() {
		prefix = VarPrefixExported
	}

	rest, ok := strings.CutPrefix(name, prefix)
	if !ok {
		return false
	}

	if rest == "" {
		return true
	}

	r, _ := utf8.DecodeRuneInString(rest)

	// https://go.dev/wiki/CodeReviewComments#mixed-caps
	return !unicode.IsLower(r)
}

func conformantType(tn *types.TypeName) bool {
	name := tn.Name()

	return strings.HasSuffix(name, TypeSuffix) ||
		strings.HasSuffix(name, TypeSuffixMultiple) &&
			errorlike.UnwrapMultipleMethod.HasMethod(tn.Type(), true)
}
