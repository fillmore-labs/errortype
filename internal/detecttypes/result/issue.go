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

package result

import (
	"go/ast"
	"go/types"
)

// AssertionIssue describes a configuration error or contradiction on a
// package-level compile-time error assertion (`var _ error = ...`).
//
//sumtype:decl
type AssertionIssue interface {
	assertionIssue()
	Target() ast.Expr
}

// UnknownDirective reports an unrecognized `//errortype:` directive.
type UnknownDirective struct {
	Expr      ast.Expr
	Directive string
}

// MismatchedDirective reports an `//errortype:` directive that disagrees with its assertion form.
type MismatchedDirective struct {
	Expr      ast.Expr
	TypeName  *types.TypeName
	Directive ErrorType
}

// ConflictingDirective reports differing valid `//errortype:` directives for the same type.
type ConflictingDirective struct {
	Expr       ast.Expr
	TypeName   *types.TypeName
	Directive  ErrorType
	Related    ast.Expr
	RelatedDir ErrorType
}

// ContradictoryAssertion reports opposing plain `var _ error` assertions for an undecided type.
type ContradictoryAssertion struct {
	Expr     ast.Expr
	TypeName *types.TypeName
	Related  ast.Expr
	FirstPtr bool
}

func (UnknownDirective) assertionIssue()       {}
func (MismatchedDirective) assertionIssue()    {}
func (ConflictingDirective) assertionIssue()   {}
func (ContradictoryAssertion) assertionIssue() {}

// Target returns the assertion expression associated with the issue.
func (i UnknownDirective) Target() ast.Expr { return i.Expr }

// Target returns the assertion expression associated with the issue.
func (i MismatchedDirective) Target() ast.Expr { return i.Expr }

// Target returns the assertion expression associated with the issue.
func (i ConflictingDirective) Target() ast.Expr { return i.Expr }

// Target returns the assertion expression associated with the issue.
func (i ContradictoryAssertion) Target() ast.Expr { return i.Expr }
