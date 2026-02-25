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

package overrides_test

import (
	"context"
	_ "embed"
	"strings"
	"testing"

	"fillmore-labs.com/errortype/internal/detecttypes/result"
	"fillmore-labs.com/errortype/internal/overrides"
)

//go:embed readtypes_test.yaml
var _yamltypes string

func TestRead_TypesMap(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ov, err := overrides.ReadTypes(ctx, strings.NewReader(_yamltypes))
	if err != nil {
		t.Fatalf("ReadTypes() error = %v", err)
	}

	if ptrs := ov.Types[result.Pointer]; len(ptrs) != 1 || ptrs[0].String() != "my/pkg.PointerErr" {
		t.Errorf("Unexpected Pointer types: %v", ptrs)
	}

	if vals := ov.Types[result.Value]; len(vals) != 1 || vals[0].String() != "my/pkg.ValueErr" {
		t.Errorf("Unexpected Value types: %v", vals)
	}

	if sups := ov.Types[result.Suppress]; len(sups) != 1 || sups[0].String() != "my/pkg.SuppressErr" {
		t.Errorf("Unexpected Suppress types: %v", sups)
	}

	// Inconsistent entries must be ignored on read; only Pointer, Value, Suppress are stored.
	if got, want := len(ov.Types), 3; got != want {
		t.Errorf("Expected 3 types (Pointer, Value, Suppress), got %d: %v", got, ov.Types)
	}
}
