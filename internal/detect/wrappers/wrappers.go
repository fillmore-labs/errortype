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

package wrappers

import (
	"go/types"
	"iter"
	"log"

	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/detect/result"
)

// LookupWrappers resolves explicitly defined wrapper functions within the package scope.
func LookupWrappers(pkgScope *types.Scope, explicitWrappers []ExplicitWrapper) result.ErrorFuncs {
	wrappers := make(result.ErrorFuncs, len(explicitWrappers))
	for _, known := range explicitWrappers {
		var fun *types.Func
		if known.Receiver == "" {
			fun, _ = pkgScope.Lookup(known.Name).(*types.Func)
		} else {
			recv := pkgScope.Lookup(known.Receiver)
			if recv == nil {
				continue
			}

			named, ok := types.Unalias(recv.Type()).(*types.Named)
			if !ok {
				log.Printf("Wrapper receiver \"%s\" is not a type", known.Receiver)

				continue
			}

			fun = findDirectMethod(named, known.Name)
		}

		if fun == nil {
			continue
		}

		wrapper, ok := findParameters(fun, known.Typ)
		if !ok {
			log.Printf("Wrapper parameters of \"%s\" can not be determined", fun.FullName())

			continue
		}

		wrappers[fun] = wrapper
	}

	return wrappers
}

// findDirectMethod returns the explicitly declared method by name on a Named type.
func findDirectMethod(named *types.Named, methodName string) *types.Func {
	for m := range allMethods(named) {
		if m.Name() == methodName {
			return m
		}
	}

	return nil
}

func allMethods(named *types.Named) iter.Seq[*types.Func] {
	// Check if the underlying type is an interface
	if iface, ok := named.Underlying().(*types.Interface); ok {
		// Iterate over the explicitly declared methods on the interface
		return iface.ExplicitMethods()
	}

	// For non-interfaces, methods are declared on the Named type itself
	return named.Methods()
}

// DetectWrappers performs graph-based AST analysis to automatically detect functions that wrap known error type checking functions.
func DetectWrappers(p *analysis.Pass, known result.ErrorFuncs) result.ErrorFuncs {
	// Phase 1: gather wrapperCandidate wrapper functions
	wrapperCandidates := findCandidates(p)

	if len(wrapperCandidates) == 0 {
		return nil
	}

	// Phase 2: verify bodies and build call graph
	wrappers := make(result.ErrorFuncs)
	for _, cand := range wrapperCandidates {
		resolveCandidate(p.TypesInfo, cand, wrapperCandidates, known, wrappers)
	}

	return wrappers
}
