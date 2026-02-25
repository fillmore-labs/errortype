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

package detectwrappers_test

import (
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	. "fillmore-labs.com/errortype/internal/detectwrappers"
	"fillmore-labs.com/errortype/internal/settings/errorwrapper"
)

func TestDetectAnalyzer(t *testing.T) {
	t.Parallel()

	dir := analysistest.TestData()
	fileName := filepath.Join(dir, "overrides.yaml")

	o := errorwrapper.DefaultConfig()
	if err := o.ReadOverrides(fileName); err != nil {
		t.Fatalf("Can't read override file: %v", err)
	}

	tests := [...]struct {
		name string
		pkg  string
	}{
		{"wrappers", "./wrappers/..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := o
			a := NewAnalyzer(&cfg)

			analysistest.Run(t, dir, a, tt.pkg)
		})
	}
}
