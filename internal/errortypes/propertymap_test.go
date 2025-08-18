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

package errortypes_test

import (
	"fmt"
	"go/token"
	"go/types"
	"testing"

	. "fillmore-labs.com/errortype/internal/errortypes"
)

type myProperty int

const (
	myFirst myProperty = 1 << iota
	mySecond
	myThird
	myNone myProperty = 0
)

func (m myProperty) DeterminedType() ErrorType {
	panic("not implemented")
}

func (m myProperty) String() string {
	return fmt.Sprintf("%03b", m)
}

func TestAddTypeProperty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		existing myProperty
		newProp  myProperty
		expected myProperty
	}{
		{
			name:     "Add to empty PropertyMap",
			existing: myNone,
			newProp:  myFirst,
			expected: myFirst,
		},
		{
			name:     "Adding same property",
			existing: myFirst,
			newProp:  myFirst,
			expected: myFirst,
		},
		{
			name:     "Combine properties",
			existing: myFirst,
			newProp:  mySecond,
			expected: myFirst | mySecond,
		},
		{
			name:     "No new properties added",
			existing: myFirst | mySecond,
			newProp:  myFirst,
			expected: myFirst | mySecond,
		},
		{
			name:     "Add multiple new properties",
			existing: myFirst | mySecond,
			newProp:  mySecond | myThird,
			expected: myFirst | mySecond | myThird,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a new PropertyMap and set the initial property
			propMap := NewPropertyMap[myProperty]()

			typeName := types.NewTypeName(token.NoPos, nil, "TestType", nil)
			if tt.existing != myNone {
				propMap.SetTypeProperty(typeName, tt.existing)
			}

			// Call AddTypeProperty
			if old := propMap.AddTypeProperty(typeName, tt.newProp); old != tt.existing {
				t.Errorf("expected old property: %v, got: %v", tt.existing, old)
			}

			// Verify results
			if actual, ok := propMap.GetTypeProperty(typeName); !ok || actual != tt.expected {
				t.Errorf("expected new property: %v, got: %v", tt.expected, actual)
			}
		})
	}
}
