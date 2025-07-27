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

package analyze

import (
	"errors"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"fillmore-labs.com/errortype/internal/detect"
)

// ErrNoInspectorResult is returned when the ast inspector is missing.
var ErrNoInspectorResult = errors.New("errortype: inspector result missing")

// ErrNoDetectTypesResult is returned when the result from the detecttypes analyzer is missing.
var ErrNoDetectTypesResult = errors.New("errortype: detecttypes result missing")

// run executes the analysis pass using the provided options. It processes detected types,
// analyzes the abstract syntax tree (AST), and calculates the final result. If any step fails,
// an error is returned. Otherwise, the computed result is returned.
func (o *options) run(ap *analysis.Pass) (any, error) {
	detectedResult, ok := ap.ResultOf[o.detecttypes].(detect.Result)
	if !ok {
		return nil, ErrNoDetectTypesResult
	}

	in, ok := ap.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, ErrNoInspectorResult
	}

	p := newPass(ap)

	p.processDetectedTypes(detectedResult.Types)

	p.processAST(in, o.styleCheck)

	res := p.calculateResult()

	return res, nil
}
