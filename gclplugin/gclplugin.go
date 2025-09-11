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

package gclplugin

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	errortype "fillmore-labs.com/errortype/analyzer"
	"fillmore-labs.com/errortype/detect"
)

func init() {
	register.Plugin("errortype", New)
}

// New creates a new [Plugin] instance with the given [Settings].
func New(rawSettings any) (register.LinterPlugin, error) {
	settings, err := register.DecodeSettings[Settings](rawSettings)
	if err != nil {
		return nil, err
	}

	return Plugin{settings: settings}, nil
}

// Plugin is the errortype linter as a [register.LinterPlugin].
type Plugin struct {
	settings Settings
}

// GetLoadMode returns the golangci load mode.
func (Plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

// BuildAnalyzers returns the [analysis.Analyzer]s for a zerolint run.
func (p Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	dopts := detectOptions(p.settings)

	d := detect.New(dopts...)

	eopts := errortypeOptions(p.settings)
	eopts = append(eopts, errortype.WithDetectTypes(d))

	e := errortype.New(eopts...)

	return []*analysis.Analyzer{e}, nil
}
