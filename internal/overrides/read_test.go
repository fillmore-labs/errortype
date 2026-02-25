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
	"os"
	"path/filepath"
	"strings"
	"testing"

	typesresult "fillmore-labs.com/errortype/internal/detecttypes/result"
	wrappersresult "fillmore-labs.com/errortype/internal/detectwrappers/result"
	"fillmore-labs.com/errortype/internal/overrides"
)

//go:embed read_test.yaml
var _yamlmixed string

func TestRead_Mixed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	typesOv, err := overrides.ReadTypes(ctx, strings.NewReader(_yamlmixed))
	if err != nil {
		t.Fatalf("ReadTypes() on mixed file error = %v", err)
	}

	if ptrs := typesOv.Types[typesresult.Pointer]; len(ptrs) != 1 || ptrs[0].String() != "my/pkg.PointerErr" {
		t.Errorf("ReadTypes() mixed got pointer %v, want my/pkg.PointerErr", ptrs)
	}

	wrappersOv, err := overrides.ReadWrappers(ctx, strings.NewReader(_yamlmixed))
	if err != nil {
		t.Fatalf("ReadWrappers() on mixed file error = %v", err)
	}

	if isFns := wrappersOv.Wrappers[wrappersresult.WrapperIs]; len(isFns) != 1 || isFns[0].String() != "my/pkg.Is" {
		t.Errorf("ReadWrappers() mixed got is wrappers %v, want my/pkg.Is", isFns)
	}
}

func TestRead_Empty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	typesOv, err := overrides.ReadTypes(ctx, strings.NewReader(""))
	if err != nil {
		t.Fatalf("ReadTypes() on empty reader error = %v", err)
	}

	if len(typesOv.Types) != 0 {
		t.Errorf("ReadTypes() on empty reader got non-empty: %v", typesOv.Types)
	}

	wrappersOv, err := overrides.ReadWrappers(ctx, strings.NewReader(""))
	if err != nil {
		t.Fatalf("ReadWrappers() on empty reader error = %v", err)
	}

	if len(wrappersOv.Wrappers) != 0 {
		t.Errorf("ReadWrappers() on empty reader got non-empty: %v", wrappersOv.Wrappers)
	}
}

func TestRead_Files(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	dir := t.TempDir()
	filePath := filepath.Join(dir, "overrides.yaml")

	if err := os.WriteFile(filePath, []byte(_yamlmixed), 0o600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	typesOv, err := overrides.ReadTypesFile(ctx, filePath)
	if err != nil {
		t.Fatalf("ReadTypesFile() error = %v", err)
	}

	if len(typesOv.Types[typesresult.Pointer]) != 1 {
		t.Errorf("ReadTypesFile() failed to read pointer types")
	}

	wrappersOv, err := overrides.ReadWrappersFile(ctx, filePath)
	if err != nil {
		t.Fatalf("ReadWrappersFile() error = %v", err)
	}

	if len(wrappersOv.Wrappers[wrappersresult.WrapperIs]) != 1 {
		t.Errorf("ReadWrappersFile() failed to read wrapper funcs")
	}

	nonExistent := filepath.Join(dir, "does_not_exist.yaml")
	if _, err := overrides.ReadTypesFile(ctx, nonExistent); err == nil {
		t.Errorf("ReadTypesFile() on non-existent file expected error, got nil")
	}

	if _, err := overrides.ReadWrappersFile(ctx, nonExistent); err == nil {
		t.Errorf("ReadWrappersFile() on non-existent file expected error, got nil")
	}
}
