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

package usage

import (
	"go/types"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// Reporter defines the interface for reporting diagnostics related to
// incorrect error type usage. Different implementations can provide
// context-specific error messages for returns or type assertions.
type Reporter interface {
	// ShouldBeValue is called when an error type that should be a value is
	// used as a pointer.
	ShouldBeValue(tn *types.TypeName)

	// ShouldBePointer is called when an error type that should be a pointer is
	// used as a value.
	ShouldBePointer(tn *types.TypeName)

	// UndeterminedUsage is called when a named error type is encountered whose
	// pointer-vs.-value usage has not been determined by the `detecttypes`
	// analyzer. This is often due to embedding the `error` interface.
	UndeterminedUsage(tn *types.TypeName, ptr bool)
}

// Check verifies that a given type `t` is used correctly (as a pointer or value)
// based on the determined or configured usage. It reports diagnostics for mismatches
// or for types with undetermined usage.
func (e ErrorUsage) Check(t types.Type, reporter Reporter) {
	if types.IsInterface(t) {
		return // We can't analyze interfaces.
	}

	// We can only analyze named types, as anonymous types ("struct{ error }")
	// cannot be configured.
	tn, ptr, ok := typeutil.TypeNameOf(t)
	if !ok {
		return
	}

	if !typeutil.PackageLevel(tn) {
		return // local type with embedded error
	}

	//  Look up the configured usage for the type.
	usage := e[tn]

	// Check the actual usage against the expected usage.
	switch usage & ExpectedMask {
	case PointerExpected:
		if !ptr {
			reporter.ShouldBePointer(tn)
		}

	case ValueExpected:
		if ptr {
			reporter.ShouldBeValue(tn)
		}

	case SuppressExpected:
		// Analysis for this type is suppressed.

	default:
		// The type's usage is not determined. This often happens when a struct
		// embeds an error type without defining its own Error() method.
		// We report this to suggest adding it to the configuration.
		reporter.UndeterminedUsage(tn, ptr)
	}

	// Record the observed.
	et := ValueObserved
	if ptr {
		et = PointerObserved
	}

	if usage&et == 0 {
		e[tn] |= et
	}
}
