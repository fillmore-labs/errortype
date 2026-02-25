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

package detectwrappers

import (
	"context"
	"log"
	"reflect"
	"slices"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/detectwrappers/result"
	"fillmore-labs.com/errortype/internal/qualname"
	"fillmore-labs.com/errortype/internal/typeutil"
)

var (
	resultType = reflect.TypeFor[result.Result]()
	factTypes  = []analysis.Fact{(*result.ErrorFunc)(nil)}
)

func extractFacts(pass *analysis.Pass) result.ErrorFuncs {
	eFuncs := make(result.ErrorFuncs)

	for _, fact := range pass.AllObjectFacts() {
		f, ok := fact.Fact.(*result.ErrorFunc)
		if !ok {
			continue
		}

		eFuncs[fact.Object] = *f
	}

	return eFuncs
}

type wrapperType struct {
	qualname.LocalFunc
	result.ErrorFunc
}

func (w wrapperType) cmp(o wrapperType) int {
	return w.Compare(o.LocalFunc)
}

// logResults logs the properties and the determined error type for each type in the PropertyMap.
func (d dpass) logResults(_ context.Context, logger *log.Logger) {
	pkgPath := typeutil.PkgPath(d.Fset, d.Pkg, d.Files)

	var wrappers []wrapperType

	for fun, wrapper := range d.ErrorFuncs {
		if fun.Pkg() != d.Pkg {
			continue
		}

		wrappers = append(wrappers, wrapperType{
			LocalFunc: qualname.FuncOf(fun).LocalFunc,
			ErrorFunc: wrapper,
		})
	}

	slices.SortFunc(wrappers, wrapperType.cmp)

	for _, wrapper := range wrappers {
		logger.Printf("%s %s: %s\n", pkgPath, wrapper.LocalFunc, wrapper.ErrorFunc)
	}
}
