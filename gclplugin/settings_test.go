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

package gclplugin_test

import (
	_ "embed"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"

	. "fillmore-labs.com/errortype/gclplugin"
)

//go:embed settings_test.yaml
var _yamlsettings string

func TestSettings(t *testing.T) {
	t.Parallel()

	for decoder := yaml.NewDecoder(strings.NewReader(_yamlsettings)); ; {
		var rawSettings any
		if err := decoder.DecodeContext(t.Context(), &rawSettings); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			t.Fatal(err)
		}

		r, err := New(rawSettings)
		if err != nil {
			t.Fatal(err)
		}

		if r.GetLoadMode() != "typesinfo" {
			t.Error("expected typesinfo load mode")
		}

		if _, err := r.BuildAnalyzers(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBrokenSettings(t *testing.T) {
	t.Parallel()

	decoder := yaml.NewDecoder(strings.NewReader("---\nunknown: false"))

	var rawSettings any
	if err := decoder.Decode(&rawSettings); err != nil {
		t.Fatal(err)
	}

	want := "unknown field"

	if _, err := New(rawSettings); err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("expected error with %q, got %v", want, err)
	}
}
