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

package detect

import (
	"context"
	"go/types"
	"log"
	"runtime/trace"
	"slices"

	"fillmore-labs.com/errortype/detect/result"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// createResult combines all determined type information into the final analyzer result.
// It merges types from the current package and dependencies (facts) and local overrides,
// with local overrides having the highest precedence.
func (p pass) createResult(ctx context.Context, eTypes result.ErrorTypes) result.Result {
	defer trace.StartRegion(ctx, "result").End()

	for tn, prop := range p.ErrorTypes {
		errorType := prop.DeterminedType()

		if p.inCurrentPkg(tn) {
			// Export this information as a fact when the type is defined in the current package.
			// These facts can then be consumed by analyzers running on packages dependent on this one.
			p.ExportObjectFact(tn, errorType.Fact())
		}

		// Add type to the result.
		// Local overrides will overwrite any existing entries from facts, only for this package.
		eTypes[tn] = errorType
	}

	// Convert map to slice for the result.
	return result.New(eTypes, p.ErrorFuncs)
}

func (p pass) extractFacts() (result.ErrorTypes, result.ErrorFuncs) {
	eTypes := make(result.ErrorTypes)
	eFuncs := make(result.ErrorFuncs)

	for _, fact := range p.AllObjectFacts() {
		switch f := fact.Fact.(type) {
		case *result.ErrorType:
			if tn, ok := fact.Object.(*types.TypeName); ok {
				eTypes[tn] = *f
			}

		case *result.ErrorFunc:
			if fn, ok := fact.Object.(*types.Func); ok {
				eFuncs[fn] = *f
			}
		}
	}

	return eTypes, eFuncs
}

type wrapperType struct {
	typeutil.LocalFuncName
	result.ErrorFunc
}

func (w wrapperType) cmp(o wrapperType) int {
	return w.Compare(o.LocalFuncName)
}

// logResults logs the properties and the determined error type for each type in the PropertyMap.
func (p pass) logResults(_ context.Context, logger *log.Logger) {
	pkgPath := typeutil.PkgPath(p.Pass)
	qf := types.RelativeTo(p.Pkg)

	for tn, errortype := range p.ErrorTypes.AllSorted {
		typeName := types.TypeString(tn.Type(), qf)
		determinedType := errortype.DeterminedType()

		logger.Printf("%s %s: %s (%s)\n", pkgPath, typeName, determinedType, errortype)
	}

	var wrappers []wrapperType

	for fun, wrapper := range p.ErrorFuncs {
		if fun.Pkg() != p.Pkg {
			continue
		}

		wrappers = append(wrappers, wrapperType{
			LocalFuncName: typeutil.FuncNameOf(fun).LocalFuncName,
			ErrorFunc:     wrapper,
		})
	}

	slices.SortFunc(wrappers, wrapperType.cmp)

	for _, wrapper := range wrappers {
		logger.Printf("%s %s: %s\n", pkgPath, wrapper.LocalFuncName, wrapper.ErrorFunc)
	}
}
