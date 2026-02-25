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

package result_test

import (
	"testing"

	"fillmore-labs.com/errortype/detect/result"
)

func TestWrapperType_UnmarshalText(t *testing.T) {
	t.Parallel()

	for i := range result.LastWrappperType + 1 {
		text := i.String()

		var got result.WrapperType
		err := got.UnmarshalText([]byte(text))

		if i == result.WrapperUnknown || i >= result.LastWrappperType {
			if err == nil {
				t.Errorf("Expected error for UnmarshalText(%q), got %v", text, err)
			}

			continue
		}

		if err != nil {
			t.Errorf("UnmarshalText(%q) error = %v", text, err)

			continue
		}

		if got != i {
			t.Errorf("UnmarshalText(%q) got = %v, want %v", text, got, i)
		}
	}
}

func TestWrapperType_MarshalText(t *testing.T) {
	t.Parallel()

	for i := range result.LastWrappperType + 1 {
		got, err := i.MarshalText()
		if i == result.WrapperUnknown || i >= result.LastWrappperType {
			if err == nil {
				t.Errorf("Expected error for MarshalText(%d), got %v", i, err)
			}

			continue
		}

		if err != nil {
			t.Errorf("MarshalText(%d) error = %v", i, err)

			continue
		}

		if want := i.String(); string(got) != want {
			t.Errorf("UnmarshalText(%d) got = %s, want %s", i, string(got), want)
		}
	}
}
