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
	"go/ast"
	"go/types"
	"iter"
	"log"

	"fillmore-labs.com/errortype/detect/result"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// LookupWrappers resolves explicitly defined wrapper functions within the package scope.
func LookupWrappers(pkg *types.Package, explicitWrappers []ExplicitWrapper, debug bool) result.ErrorFuncs {
	scope := pkg.Scope()

	wrappers := make(result.ErrorFuncs, len(explicitWrappers))
	for _, known := range explicitWrappers {
		var (
			fun types.Object
			sig *types.Signature
		)

		if known.Receiver == "" {
			fun = scope.Lookup(known.Name)
			if fun == nil {
				if debug {
					log.Printf("%s %s: not found in this pass", pkg.Path(), known.Name)
				}

				continue // not in this pass; may resolve in a test variant
			}

			sig = typeutil.SignatureOf(fun)
			if sig == nil {
				log.Printf("%s %s: is not a function or function-typed variable", pkg.Path(), known.Name)
				continue
			}
		} else {
			recv := scope.Lookup(known.Receiver)
			if recv == nil {
				continue
			}

			named, ok := types.Unalias(recv.Type()).(*types.Named)
			if !ok {
				log.Printf("%s %s: is not a type", pkg.Path(), known.Receiver)

				continue
			}

			method := findDirectMethod(named, known.Name)
			if method == nil {
				continue
			}

			fun, sig = method, method.Signature()
		}

		wrapper, ok := findParameters(sig, known.Type)
		if !ok {
			log.Printf("%s %s: parameters can not be determined", pkg.Path(), known.LocalFuncName)

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
func DetectWrappers(info *types.Info, files []*ast.File, known result.ErrorFuncs) result.ErrorFuncs {
	// candidate wrapper functions
	wrapperCandidates := findCandidates(info, files, known)

	if len(wrapperCandidates) == 0 {
		return nil
	}

	// verify bodies and build call graph
	wrappers := resolveCandidates(info, wrapperCandidates, known)

	return wrappers
}
