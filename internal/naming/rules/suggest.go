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
	"fmt"
	"go/types"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Suggest proposes a conforming name for the given error variable or type object.
// It returns the zero [Suggestion], with an empty [Suggestion.Name], if the
// object's kind is unsupported.
func Suggest(obj types.Object) Suggestion {
	switch def := obj.(type) {
	case *types.Const, *types.Var:
		return suggestVar(def)

	case *types.TypeName:
		return suggestType(def)

	default:
		return Suggestion{}
	}
}

// Suggestion is a proposed name, split where [Suggestion.Numbered] inserts its
// count: for variables the prefix is the whole name (“errFoo” + “”), for types
// the suffix is [TypeSuffix] (“foo” + “Error”).
type Suggestion struct {
	prefix, suffix string
}

// Name returns the suggested name.
func (s Suggestion) Name() string {
	return s.prefix + s.suffix
}

// Numbered returns the suggested name with the count 'i' inserted between prefix
// and suffix, e.g. “errFoo2” or “foo2Error”.
func (s Suggestion) Numbered(i int) string {
	p := s.prefix
	if p == "" {
		p = "E" // can only happen for an "Error" type suggestion
	}

	return fmt.Sprintf("%s%d%s", p, i, s.suffix)
}

func suggestVar(obj types.Object) Suggestion {
	name := obj.Name()
	if name == "e" { // special case: "e" should be "err"
		return Suggestion{VarPrefix, ""}
	}

	prefix := VarPrefix
	if obj.Exported() {
		prefix = VarPrefixExported
	}

	stem := trimErr(name)
	if stem == "" {
		return Suggestion{prefix, ""}
	}

	first, size := utf8.DecodeRuneInString(stem)
	stem = stem[size:]

	first = unicode.ToUpper(first)

	var sb strings.Builder
	sb.Grow(len(prefix) + utf8.RuneLen(first) + len(stem))

	_, _ = sb.WriteString(prefix) // ignore error
	_, _ = sb.WriteRune(first)    // ignore error
	_, _ = sb.WriteString(stem)   // ignore error

	return Suggestion{sb.String(), ""}
}

func suggestType(def *types.TypeName) Suggestion {
	stem := trimErr(def.Name())
	if stem == "" {
		p := ""
		if !def.Exported() {
			p = "e"
		}

		return Suggestion{p, TypeSuffix}
	}

	first, size := utf8.DecodeRuneInString(stem)
	stem = stem[size:]

	var prefix string

	switch {
	case !unicode.IsLetter(first):
		prefix = "e"
		if def.Exported() {
			prefix = "E"
		}

	case def.Exported():
		first = unicode.ToUpper(first)

	default:
		first = unicode.ToLower(first)
	}

	var sb strings.Builder
	sb.Grow(len(prefix) + utf8.RuneLen(first) + len(stem))

	_, _ = sb.WriteString(prefix) // ignore error
	_, _ = sb.WriteRune(first)    // ignore error
	_, _ = sb.WriteString(stem)   // ignore error

	return Suggestion{sb.String(), TypeSuffix}
}

// Longest first, so that e.g. “errorsFoo” trims “errors” rather than “error”.
var trim = [...]string{"errors", TypeSuffixMultiple, "error", TypeSuffix, VarPrefix, VarPrefixExported}

func trimErr(name string) string {
	for _, prefix := range trim {
		if s, ok := strings.CutPrefix(name, prefix); ok {
			name = s
			break
		}
	}

	for _, suffix := range trim {
		if s, ok := strings.CutSuffix(name, suffix); ok {
			name = s
			break
		}
	}

	return name
}
