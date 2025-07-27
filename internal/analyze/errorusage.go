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

package analyze

import (
	"go/types"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// checkErrorUsage verifies that a given type `t` is used correctly (as a pointer or value)
// based on the determined or configured usage. It reports diagnostics for mismatches
// or for types with undetermined usage.
func (p pass) checkErrorUsage(t types.Type, reporter UsageReporter) {
	if types.IsInterface(t) {
		return // We can't analyze interfaces.
	}

	// We can only analyze named types, as anonymous types ("struct{ error }")
	// cannot be configured.
	tn, isPtr, ok := typeutil.TypeNameOf(t)
	if !ok {
		return
	}

	if tn.Parent() != tn.Pkg().Scope() {
		return // local type with embedded error
	}

	// Record the observed usage and look up the expected one.
	usage := p.recordAndLookup(tn, isPtr)

	// Check the actual usage against the expected usage.
	switch usage {
	case PointerExpected:
		if !isPtr {
			reporter.ShouldBePointer(tn)
		}

	case ValueExpected:
		if isPtr {
			reporter.ShouldBeValue(tn)
		}

	case SuppressExpected:
		// Analysis for this type is suppressed.

	case None:
		// The type's usage is not determined. This often happens when a struct
		// embeds an error type without defining its own Error() method.
		// We report this to suggest adding it to the configuration.
		reporter.UndeterminedUsage(tn, isPtr)

	default:
		// This should not happen if the analyzer is configured correctly.
		panic("Misconfigured type in usage map: " + tn.Name())
	}
}

// recordAndLookup records the observed usage type (value or pointer) for the given
// type name and returns the detected usage for that type, masked by AnalyzeMask.
// It updates the errorUsages property with the observed usage before returning the result.
func (p pass) recordAndLookup(tn *types.TypeName, isPtr bool) Usage {
	// Record the observed ...
	et := ValueObserved
	if isPtr {
		et = PointerObserved
	}

	// ... and look up the configured usage for the type.
	return p.errorUsages.AddTypeProperty(tn, et) & ExpectedMask
}
