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

package usage_test

import (
	"go/token"
	"go/types"
	"testing"

	. "fillmore-labs.com/errortype/internal/analyze/usage"
)

func TestCheck(t *testing.T) {
	t.Parallel()

	type ExpectedReport int

	const (
		_ ExpectedReport = iota // No Report expected
		ShouldBePointer
		ShouldBeValue
		UndeterminedPointer
		UndeterminedValue
	)

	// Set up shared types for test cases
	pkg := types.NewPackage("test", "test")

	errorInterface := types.Universe.Lookup("error").Type()
	structType := types.NewStruct([]*types.Var{types.NewField(token.NoPos, pkg, "error", errorInterface, true)}, nil)

	tn := types.NewTypeName(token.NoPos, pkg, "MyError", nil)
	pkg.Scope().Insert(tn)

	namedType := types.NewNamed(tn, structType, nil)

	ln := types.NewTypeName(token.NoPos, pkg, "MyLocalError", nil)
	local := types.NewScope(pkg.Scope(), token.NoPos, token.NoPos, "")
	local.Insert(ln)

	localType := types.NewNamed(ln, structType, nil)

	ptrType := types.NewPointer(namedType)

	testCases := [...]struct {
		checkType       types.Type
		tn              *types.TypeName
		name            string
		wantReport      ExpectedReport
		initialProperty Usage
		finalProperty   Usage
	}{
		// Early exit cases (no reporting expected)
		{
			name:      "InterfaceType",
			checkType: errorInterface,
		},
		{
			name:      "AnonymousType",
			checkType: structType,
		},
		{
			name:      "LocalType",
			checkType: localType,
			tn:        ln,
		},

		// Expected pointer usage cases
		{
			name:            "PointerExpected_UsedAsPointer",
			checkType:       ptrType,
			tn:              tn,
			initialProperty: PointerExpected,
			finalProperty:   PointerExpected | PointerObserved,
		},
		{
			name:            "PointerExpected_UsedAsValue",
			checkType:       namedType,
			tn:              tn,
			initialProperty: PointerExpected,
			finalProperty:   PointerExpected | ValueObserved,
			wantReport:      ShouldBePointer,
		},

		// Expected value usage cases
		{
			name:            "ValueExpected_UsedAsValue",
			checkType:       namedType,
			tn:              tn,
			initialProperty: ValueExpected,
			finalProperty:   ValueExpected | ValueObserved,
		},
		{
			name:            "ValueExpected_UsedAsPointer",
			checkType:       ptrType,
			tn:              tn,
			initialProperty: ValueExpected,
			finalProperty:   ValueExpected | PointerObserved,
			wantReport:      ShouldBeValue,
		},

		// Expected suppressed diagnostic cases
		{
			name:            "SuppressExpected_UsedAsPointer",
			checkType:       ptrType,
			tn:              tn,
			initialProperty: SuppressExpected,
			finalProperty:   SuppressExpected | PointerObserved,
		},
		{
			name:            "SuppressExpected_UsedAsValue",
			checkType:       namedType,
			tn:              tn,
			initialProperty: SuppressExpected,
			finalProperty:   SuppressExpected | ValueObserved,
		},

		// Undetermined usage cases
		{
			name:          "UndeterminedUsage_Pointer",
			checkType:     ptrType,
			tn:            tn,
			finalProperty: PointerObserved,
			wantReport:    UndeterminedPointer,
		},
		{
			name:          "UndeterminedUsage_Value",
			checkType:     namedType,
			tn:            tn,
			finalProperty: ValueObserved,
			wantReport:    UndeterminedValue,
		},

		// Usage accumulation cases
		{
			name:            "PointerExpected_AccumulatesValueObserved",
			checkType:       namedType,
			tn:              tn,
			initialProperty: PointerExpected | PointerObserved,
			finalProperty:   PointerExpected | PointerObserved | ValueObserved,
			wantReport:      ShouldBePointer,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			errorUsage := make(ErrorUsage)
			if tc.tn != nil {
				errorUsage[tc.tn] = tc.initialProperty
			}

			reporter := &moqReporter{}

			switch tc.wantReport {
			case ShouldBeValue:
				reporter.ShouldBeValueFunc = func(*types.TypeName) {}

			case ShouldBePointer:
				reporter.ShouldBePointerFunc = func(*types.TypeName) {}

			case UndeterminedPointer, UndeterminedValue:
				reporter.UndeterminedUsageFunc = func(*types.TypeName, bool) {}
			}

			func() {
				// MockReporter panics on undefined calls
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Unexpected report method called: %v", r)
					}
				}()

				errorUsage.Check(tc.checkType, reporter)
			}()

			// Check reporter calls
			switch tc.wantReport {
			case ShouldBeValue:
				shouldBeValueCalls := reporter.ShouldBeValueCalls()

				if len(shouldBeValueCalls) != 1 {
					t.Fatalf("Expected 1 ShouldBeValue call, got %d", len(shouldBeValueCalls))
				}

				if shouldBeValueCalls[0].Tn != tc.tn {
					t.Errorf("Expected type name %v, got %v", tc.tn, shouldBeValueCalls[0].Tn)
				}

			case ShouldBePointer:
				shouldBePointerCalls := reporter.ShouldBePointerCalls()

				if len(shouldBePointerCalls) != 1 {
					t.Fatalf("Expected 1 ShouldBePointer call, got %d", len(shouldBePointerCalls))
				}

				if shouldBePointerCalls[0].Tn != tc.tn {
					t.Errorf("Expected type name %v, got %v", tc.tn, shouldBePointerCalls[0].Tn)
				}

			case UndeterminedPointer, UndeterminedValue:
				undeterminedUsageCalls := reporter.UndeterminedUsageCalls()

				if len(undeterminedUsageCalls) != 1 {
					t.Fatalf("Expected 1 UndeterminedUsage call, got %d", len(undeterminedUsageCalls))
				}

				call := undeterminedUsageCalls[0]
				if call.Tn != tc.tn {
					t.Errorf("Expected type name %v, got %v", tc.tn, call.Tn)
				}

				if wantPtr := tc.wantReport == UndeterminedPointer; call.Ptr != wantPtr {
					t.Errorf("Expected ptr=%v, got %v", wantPtr, call.Ptr)
				}
			}

			// Check final property state
			if tc.tn != nil {
				u, ok := errorUsage[tc.tn]
				if !ok {
					t.Fatalf("Expected type property to be set for %s", tc.tn.Name())
				}

				if u != tc.finalProperty {
					t.Errorf("Expected final property %v, got %v", tc.finalProperty, u)
				}
			}
		})
	}
}
