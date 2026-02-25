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

// Package assertions tests contradictory assertions and hybrid local overrides.
package assertions

import "test/assertions/dep"

type (
	localConflictError      struct{ error }
	localMultiConflictError struct{ error }
)

// Local contradictory undecided assertions (second flagged with Related to first).
var (
	_ error = localConflictError{}       // want ` \(et:emb\)$`
	_ error = (*localConflictError)(nil) // want ` \(et:emb\+\)$` ` \(et:cfl\)$`
)

// Multi-name contradictory spec.
var _, _ error = (*localMultiConflictError)(nil), // want ` \(et:emb\+\)$`
	localMultiConflictError{} // want ` \(et:emb\)$` ` \(et:cfl\)$`

// Imported undecided tie-breaker, plus decided agreeing and disagreeing assertions.
var (
	_ error = (*dep.UndecidedError)(nil)
	_ error = dep.AgreedValError{}
	_ error = (*dep.DisagreedValError)(nil) // want ` \(et:var\)$`
)

// Imported directive overrides (win even when upstream disagrees; suppress allows both).
var (
	//errortype:pointer
	_ error = (*dep.OverridePtrError)(nil)

	//errortype:value
	_ error = dep.OverrideValError{}

	//errortype:suppress
	_ error = (*dep.OverrideSuppressError)(nil)
	_ error = dep.OverrideSuppressError{}
)

// Mismatched, unknown, and conflicting directives.
var (
	//errortype:pointer
	_ error = dep.MismatchDirValError{} // want ` \(et:dir\)$`

	//errortype:value
	_ error = (*dep.MismatchDirPtrError)(nil) // want ` \(et:var\)$` ` \(et:dir\)$`

	//errortype:bogus
	_ error = dep.UnknownDirectiveError{} // want ` \(et:dir\)$`

	//errortype:pointer
	_ error = (*dep.ConflictDirectiveError)(nil) // want ` \(et:var\)$`
	//errortype:value
	_ error = dep.ConflictDirectiveError{} // want ` \(et:dir\)$`
)

func DownstreamUsage(choice int) error {
	switch choice {
	case 0:
		return dep.UndecidedError{} // want ` \(et:ret\+\)$`
	case 1:
		return &dep.UndecidedError{}
	case 2:
		return dep.OverridePtrError{} // want ` \(et:ret\+\)$`
	case 3:
		return &dep.OverrideValError{} // want ` \(et:ret\)$`
	case 4:
		return dep.OverrideSuppressError{}
	case 5:
		return &dep.OverrideSuppressError{}
	default:
		return nil
	}
}
