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

	"fillmore-labs.com/errortype/detect/result"
	"fillmore-labs.com/errortype/internal/overrides"
)

//go:embed read_test.yaml
var _yamloverrides string

func TestRead_WrappersMap(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ov, err := overrides.Read(ctx, strings.NewReader(_yamloverrides))
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if len(ov.Wrappers) != 3 {
		t.Errorf("Expected 3 wrapper types, got %d", len(ov.Wrappers))
	}

	if fns := ov.Wrappers[result.WrapperIs]; len(fns) != 1 || fns[0].String() != "my/pkg.Is" {
		t.Errorf("Unexpected WrapperIs funcs: %v", fns)
	}

	if fns := ov.Wrappers[result.FuncIsType]; len(fns) != 1 || fns[0].String() != "my/pkg.IsType" {
		t.Errorf("Unexpected FuncIsType funcs: %v", fns)
	}

	if fns := ov.Wrappers[result.FuncEqual]; len(fns) != 1 || fns[0].String() != "my/pkg.Equal" {
		t.Errorf("Unexpected FuncEqual funcs: %v", fns)
	}
}
