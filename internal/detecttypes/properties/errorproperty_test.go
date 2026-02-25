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

package properties_test

import (
	"strings"
	"testing"

	. "fillmore-labs.com/errortype/internal/detecttypes/properties"
)

func TestErrorProperty_String(t *testing.T) {
	t.Parallel()

	actual := (ErrorProperty(1)<<NumProperties - 1).String()

	seen := make(map[string]struct{}, NumProperties)
	for word := range strings.SplitSeq(actual, ", ") {
		seen[word] = struct{}{}
	}

	if got, want := len(seen), NumProperties; got != want {
		t.Errorf("Expected %s to have %d distinct components, but got %d", actual, want, got)
	}
}
