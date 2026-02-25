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

package run

import (
	"context"
	"errors"
	"fmt"
	"runtime/trace"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"fillmore-labs.com/errortype/detect/result"
	"fillmore-labs.com/errortype/internal/analyze"
	"fillmore-labs.com/errortype/internal/naming"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// ErrResultMissing is returned when the result from an analyzer is missing.
var ErrResultMissing = errors.New("analyzer result missing")

// Run executes the analysis pass using the provided options. It processes detected types,
// analyzes the abstract syntax tree (AST), and calculates the final result. If any step fails,
// an error is returned. Otherwise, the computed result is returned.
func (o *Options) Run(ap *analysis.Pass) (any, error) {
	in, ok := ap.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, fmt.Errorf("errortype: %s: %w", inspect.Analyzer.Name, ErrResultMissing)
	}

	detectedResult, ok := ap.ResultOf[o.DetectTypes].(result.Result)
	if !ok {
		return nil, fmt.Errorf("errortype: %s: %w", o.DetectTypes.Name, ErrResultMissing)
	}

	ctx := context.Background()

	ctx, task := trace.NewTask(ctx, "errortype")
	defer task.End()

	if trace.IsEnabled() {
		trace.Log(ctx, "pkg", typeutil.PkgPath(ap))
	}

	if o.Naming() {
		naming.CheckNaming(ctx, ap)
	}

	p := analyze.NewPass(ap, detectedResult, o.Options)

	p.ProcessAST(ctx, in)

	if o.Suggest != "" {
		suggestions := p.Suggestions(ctx)

		pkgPath := typeutil.PkgPath(p.Pass)

		if err := o.writeSuggestions(ctx, suggestions, pkgPath); err != nil {
			return nil, err
		}
	}

	return nil, nil
}
