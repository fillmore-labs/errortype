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
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/errortypes"
)

// pass holds the state for a single run of the detecttypes analyzer on a package.
// It embeds the underlying *analysis.Pass to provide direct access to its fields
// and methods, and it holds the map of properties discovered for types defined
// within the current package.
type pass struct {
	*analysis.Pass
	errortypes.PropertyMap[ErrorProperty]
	StyleCheck bool
}

// newPass creates and initializes a new pass for the detecttypes analyzer.
func newPass(ap *analysis.Pass) pass {
	return pass{
		Pass:        ap,
		PropertyMap: errortypes.NewPropertyMap[ErrorProperty](),
	}
}

// AllTypeDecls is an iterator over all type specifications (*ast.TypeSpec) in the pass's files.
func (p pass) AllTypeDecls(yield func(*ast.TypeSpec) bool) {
	iterateOverSpecs(p.Files, token.TYPE, yield)
}

// AllVarDecls is an iterator over all variable value specifications (*ast.ValueSpec) in the pass's files.
func (p pass) AllVarDecls(yield func(*ast.ValueSpec) bool) {
	iterateOverSpecs(p.Files, token.VAR, yield)
}

// AllFuncDecls is an iterator over all function declarations (*ast.FuncDecl) in the pass's files.
func (p pass) AllFuncDecls(yield func(*ast.FuncDecl) bool) {
	iterateOverDecls(p.Files, yield)
}
