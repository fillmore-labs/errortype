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

package detect

import (
	"fmt"
	"go/ast"
	"go/types"
	"log"
	"strings"

	"fillmore-labs.com/errortype/internal/errortypes"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// LogErrorf reports an internal ("should not happen") failure message.
func (p pass) LogErrorf(n ast.Node, format string, args ...any) {
	var sb strings.Builder
	_, _ = sb.WriteString("Internal error: ")
	_, _ = fmt.Fprintf(&sb, format, args...)
	_ = sb.WriteByte('\n')

	_ = ast.Fprint(&sb, p.Fset, n, ast.NotNilFilter)

	log.Println(sb.String())
}

// logResults logs the determined error types for each type name in the PropertyMap.
// It includes additional information if there is an error in the determined type.
func (p pass) logResults() {
	qf := types.RelativeTo(p.Pkg)

	for tn, errortype := range p.AllSorted {
		determinedType := errortype.DeterminedType()

		var extra string
		if mismatch := determinedTypeCheck(tn, determinedType); mismatch != "" {
			extra = " !!! " + mismatch + " !!!"
		}

		log.Printf("%s %s: %s (%s)%s", p.Pkg.Path(), types.TypeString(tn.Type(), qf), determinedType, errortype, extra)
	}
}

// determinedTypeCheck verifies if the determined error type matches the possible uses.
// It returns a string describing any mismatch or an empty string if there are no issues.
func determinedTypeCheck(tn *types.TypeName, determinedType errortypes.ErrorType) string {
	switch determinedType {
	case errortypes.PointerType:
		if !typeutil.HasErrorMethod(types.NewPointer(tn.Type())) {
			return "missing pointer error method"
		}

	case errortypes.ValueType:
		if !typeutil.HasErrorMethod(tn.Type()) {
			return "missing value error method"
		}

	case errortypes.Undecided:
		if !typeutil.HasErrorMethod(tn.Type()) {
			return "missing value error method"
		} else if !typeutil.HasErrorMethod(types.NewPointer(tn.Type())) {
			return "missing pointer error method"
		}
	}

	return ""
}
