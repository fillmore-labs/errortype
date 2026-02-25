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

	"fillmore-labs.com/errortype/internal/naming/plan"
)

// Report emits one [analysis.Diagnostic] per rename candidate.
//
// A [analysis.SuggestedFix] is attached only when [plan.Candidate.Fixable] is true;
// when [plan.Candidate.Deprecate] is true, the fix also inserts a "Deprecated:"
// alias redirecting to the new name. If customMessage is non-nil, it overrides
// [DefaultMessage].
func Report(pass *analysis.Pass, renames []plan.Candidate, customMessage MessageFunc) {
	for _, rn := range renames {
		report(pass, rn, customMessage)
	}
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
		fixes = createFixes(pass.Fset, rn)
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
