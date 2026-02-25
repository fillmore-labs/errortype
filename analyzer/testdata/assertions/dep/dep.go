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

// Package dep provides upstream error types for assertion and directive tests.
package dep

type (
	UndecidedError         struct{ error }
	AgreedValError         struct{}
	DisagreedValError      struct{}
	OverridePtrError       struct{}
	OverrideValError       struct{}
	OverrideSuppressError  struct{}
	MismatchDirValError    struct{}
	MismatchDirPtrError    struct{}
	ConflictDirectiveError struct{}
	UnknownDirectiveError  struct{}
)

func (AgreedValError) Error() string         { return "" }
func (DisagreedValError) Error() string      { return "" }
func (OverridePtrError) Error() string       { return "" }
func (OverrideValError) Error() string       { return "" }
func (OverrideSuppressError) Error() string  { return "" }
func (MismatchDirValError) Error() string    { return "" }
func (MismatchDirPtrError) Error() string    { return "" }
func (ConflictDirectiveError) Error() string { return "" }
func (UnknownDirectiveError) Error() string  { return "" }

var (
	_ error = AgreedValError{}
	_ error = DisagreedValError{}
	_ error = OverridePtrError{}
	_ error = (*OverrideValError)(nil)
	_ error = OverrideSuppressError{}
	_ error = MismatchDirValError{}
	_ error = MismatchDirPtrError{}
	_ error = ConflictDirectiveError{}
	_ error = UnknownDirectiveError{}
)
