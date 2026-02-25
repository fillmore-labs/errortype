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

package c

// Usage through a local alias is attributed to the aliased type, and the alias mirrors its decision.
type (
	AliasUsedPtrError  struct{ error }      // want AliasUsedPtrError:"pointer"
	AliasUsedPtrAlias  = AliasUsedPtrError  // want AliasUsedPtrAlias:"pointer"
	AliasUsedValError  struct{ error }      // want AliasUsedValError:"value"
	AliasUsedValAlias  = AliasUsedValError  // want AliasUsedValAlias:"value"
	PtrAliasUsedError  struct{ error }      // want PtrAliasUsedError:"pointer"
	PtrAliasUsedAlias  = *PtrAliasUsedError // want PtrAliasUsedAlias:"value"
	AliasConflictError struct{ error }      // want AliasConflictError:"undecided"
	AliasConflictAlias = AliasConflictError // want AliasConflictAlias:"undecided"
	AliasOverrideError struct{ error }      // want AliasOverrideError:"pointer"
	AliasOverrideAlias = AliasOverrideError // want AliasOverrideAlias:"pointer"
)

func _() error { return &AliasUsedPtrAlias{} }

func _() error { return AliasUsedValAlias{} }

func _(p PtrAliasUsedAlias) error { return p }

func _() error { return AliasConflictError{} }

func _() error { return &AliasConflictAlias{} }

// The override on the aliased type applies to the alias.
func _() error { return AliasOverrideAlias{} }
