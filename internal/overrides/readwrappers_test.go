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

	"fillmore-labs.com/errortype/internal/detectwrappers/result"
	"fillmore-labs.com/errortype/internal/overrides"
)

//go:embed readwrappers_test.yaml
var _yamlwrappers string

func TestRead_WrappersMap(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ov, err := overrides.ReadWrappers(ctx, strings.NewReader(_yamlwrappers))
	if err != nil {
		t.Fatalf("ReadWrappers() error = %v", err)
	}

	if got, want := len(ov.Wrappers), 3; got != want {
		t.Errorf("ReadWrappers() got %d wrapper types, want %d", got, want)
	}

	if fns := ov.Wrappers[result.WrapperIs]; len(fns) != 1 || fns[0].String() != "my/pkg.Is" {
		t.Errorf("Unexpected WrapperIs funcs: %v", fns)
	}

	if fns := ov.Wrappers[result.WrapperAsType]; len(fns) != 1 || fns[0].String() != "my/pkg.AsType" {
		t.Errorf("Unexpected WrapperAsType funcs: %v", fns)
	}

	if fns := ov.Wrappers[result.WrapperErrorf]; len(fns) != 1 || fns[0].String() != "(my/pkg.Wrapper).Errorf" {
		t.Errorf("Unexpected WrapperErrorf funcs: %v", fns)
	}
}

func TestRead_WrappersUnsupportedType(t *testing.T) {
	t.Parallel()

	const yamlwrappers = "---\nwrappers:\n  istype:\n    - my/pkg.IsType\n"

	want := `unknown wrapper type: "istype"`

	if _, err := overrides.ReadWrappers(t.Context(), strings.NewReader(yamlwrappers)); err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("Expected error containing %q, got %v", want, err)
	}
}
