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
	"strings"
	"testing"

	"github.com/goccy/go-yaml"

	. "fillmore-labs.com/errortype/gclplugin"
)

//go:embed settings_test.yaml
var _yamlsettings string

func TestSettings(t *testing.T) {
	t.Parallel()

	rawSettings := decodeSettings(t, _yamlsettings)

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

func TestBrokenSettings(t *testing.T) {
	t.Parallel()

	rawSettings := decodeSettings(t, "---\nunknown: false")

	want := "unknown field"

	if _, err := New(rawSettings); err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("expected error with %q, got %v", want, err)
	}
}

func decodeSettings(tb testing.TB, ys string) any {
	tb.Helper()

	var rawSettings any
	if err := yaml.NewDecoder(strings.NewReader(ys)).Decode(&rawSettings); err != nil {
		tb.Fatal(err)
	}

	return rawSettings
}
