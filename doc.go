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

/*
errortype is a static analysis tool that helps prevent subtle bugs in error handling.

Usage:

	errortype [flags] [packages]

It performs three checks:

 1. Inconsistent Error Type Usage: Ensures error types are used consistently
    as either pointers or values in returns, type assertions, and errors.As calls.

 2. Pointless Comparisons: Detects comparisons against newly allocated addresses
    (like errors.Is(err, &url.Error{}) or ptr == &MyStruct{}), which are almost always incorrect.

 3. Error Naming Conventions (opt-in): Checks that sentinel error variables use the Err prefix (e.g.,
    ErrNotFound) and structured error types use the Error suffix (e.g., ParseError).

For inconsistent error type usage, it automatically determines the correct usage
for most error types. Ambiguous cases can be resolved through an override file;
see the -overrides and -suggest flags.

The flags are:

	-c int
		display offending line with this many lines of context (default -1)
	-check-is
		suppress compare diagnostic on errors.Is if the type has an "Is(error) bool" or "Unwrap() error" method (default true)
	-check-unused
		report unchecked calls on errors.As-like functions (default true)
	-deep-is-check
		diagnose all "Unwrap" functions in "Is" methods, not only those on target
	-fix
		apply all suggested fixes
	-heuristics list
		list of heuristics used; values: "var", "usage", "receivers", "off" to disable (default all)
	-naming
		check error sentinel and type names
	-overrides file
		read error type overrides from this file
	-prefix-filter
		restrict variable analysis to variables with an "err" prefix (default true)
	-style-check
		check for confusing uses of errors.As
	-suggest file
		append suggestions to this file, "-" for standard output
	-test
		indicates whether test files should be analyzed, too (default true)
	-tracetypes regex
		information of error type detection in packages matching this regex
	-unchecked-assert
		report unchecked type asserts on errors

# Examples

To check the current package:

	errortype .

To check errors across packages, using and suggesting overrides:

	errortype -overrides=overrides.yaml -suggest=- ./...
*/
package main
