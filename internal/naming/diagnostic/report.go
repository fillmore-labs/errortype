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

package diagnostic

import (
	"go/token"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/astutil"
	"fillmore-labs.com/errortype/internal/naming/plan"
	"fillmore-labs.com/errortype/internal/naming/resolve"
	"fillmore-labs.com/errortype/internal/naming/rules"
)

// ReportViolations emits one [analysis.Diagnostic] per violation.
//
// It uses [plan.Plan] to map use sites, classifies embeddings, and build one
// [plan.Candidate] per violation, proposing conflict-free names and marking
// renames non-fixable where conflicts arise.
//
// A [analysis.SuggestedFix] is attached only when [plan.Candidate.Fixable] is true;
// for exported declarations the fix also inserts a "Deprecated:" alias redirecting
// to the new name. If customMessage is non-nil, it overrides [DefaultMessage].
func ReportViolations(pass *analysis.Pass, fileMap astutil.FileMap, violations []rules.Violation, customMessage MessageFunc) error {
	if len(violations) == 0 {
		return nil
	}

	resolver := resolve.New(pass.TypesInfo, pass.Pkg.Scope(), pass.Files, len(violations))

	renames, err := plan.Plan(pass.TypesInfo, resolver, fileMap, violations)
	if err != nil {
		return err
	}

	for _, rn := range renames {
		report(pass, rn, customMessage)
	}

	return nil
}

const category = "errorname"

func report(pass *analysis.Pass, rn plan.Candidate, customMessage MessageFunc) {
	obj, newName := rn.Obj, rn.NewName

	var msg string
	if customMessage != nil {
		msg = customMessage(obj, newName)
	} else {
		msg = DefaultMessage(obj, newName)
	}

	if msg == "" {
		return
	}

	var fixes []analysis.SuggestedFix
	if rn.Fixable {
		fixes = createFixes(pass.Fset, obj, newName, rn.Uses, rn.DocComment, rn.Deprecate)
	}

	start := obj.Pos()
	end := token.Pos(int(start) + len(obj.Name()))
	pass.Report(analysis.Diagnostic{
		Pos:            start,
		End:            end,
		Category:       category,
		Message:        msg,
		SuggestedFixes: fixes,
	})
}
