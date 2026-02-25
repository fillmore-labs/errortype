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

package naming

import (
	"context"
	"go/ast"
	"go/token"
	"runtime/trace"
	"strings"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/typeutil"
)

// CheckNaming helps with [error naming].
//
// [error naming]: https://go.dev/wiki/Errors#naming
func CheckNaming(ctx context.Context, p *analysis.Pass) {
	defer trace.StartRegion(ctx, "checkNaming").End()

	for decl := range typeutil.AllGenDecls(p.Files) {
		switch decl.Tok {
		case token.CONST, token.VAR:
			for spec := range typeutil.AllSpecs[*ast.ValueSpec](decl) {
				checkVarNaming(p, spec)
			}

		case token.TYPE:
			for spec := range typeutil.AllSpecs[*ast.TypeSpec](decl) {
				checkTypeNaming(p, spec)
			}
		}
	}
}

func checkVarNaming(p *analysis.Pass, spec *ast.ValueSpec) {
	for _, name := range spec.Names {
		if name.Name == "_" {
			continue
		}

		typ := p.TypesInfo.Defs[name]
		if typ == nil {
			continue
		}

		if !typeutil.HasErrorMethod(typ.Type()) {
			continue
		}

		wanted := "err"
		if name.IsExported() {
			wanted = "Err"
		}

		if strings.HasPrefix(name.Name, wanted) {
			continue
		}

		p.ReportRangef(name, "Error sentinel %q should start with %q (et:nam)", name.Name, wanted)
	}
}

func checkTypeNaming(p *analysis.Pass, spec *ast.TypeSpec) {
	name := spec.Name
	if name.Name == "_" {
		return
	}

	typ := p.TypesInfo.Defs[name]
	if typ == nil { // should not happen
		return
	}

	// We could filter out inferfaces with typ.Type().Underlying().(*types.Interface)
	// but naming them "...Error" seems like a good idea

	if !typeutil.HasErrorMethod(typ.Type()) {
		return
	}

	const wanted = "Error"

	if strings.HasSuffix(name.Name, wanted) {
		return
	}

	if strings.HasSuffix(name.Name, "Errors") && typeutil.HasMethod(typ.Type(), "Unwrap", typeutil.HasUnwrapMultipleSig) {
		return
	}

	p.ReportRangef(name, "Error type %q should end with %q (et:nam+)", name.Name, wanted)
}
