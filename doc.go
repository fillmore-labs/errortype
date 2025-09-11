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
errortype is a Go static analysis tool that helps prevent subtle bugs in error handling.

Usage:

	errortype [flags] [package ...]

It performs two main checks:

 1. Inconsistent Error Type Usage: It analyzes function returns, type assertions,
    and calls to functions like errors.As to ensure that custom error types
    are used consistently as either pointers or values.

 2. Pointless Pointer Comparisons: It detects comparisons of pointers against
    the address of a newly created value (e.g., 'ptr == &MyStruct{}'), which
    are almost always incorrect.

For inconsistent error type usage, it automatically determines the correct usage
for most error types but may require a configuration file for ambiguous cases.

The flags are:

		-c int
		  	display offending line with this many lines of context (default -1)
		-check-is
		  	suppress compare diagnostic on errors.Is if the compared type has an "Is(error) bool" method (default true)
		-deep-is-check
		  	diagnose all "Unwrap" functions in "Is" methods, not only on target (default false)
	  	-unchecked-assert
	    	report unchecked type asserts on errors (default false)
		-heuristics value
		  	list of heuristics used (default: "usage,receivers", "off" to disable)
		-overrides value
		  	read error type overrides from this file
		-style-check
		  	check for confusing uses of errors.As (default true)
		-suggest string
		  	append override suggestions to this file, "-" for standard output
		-tags string
		  	comma-separated list of build tags to apply
		-test
		  	indicates whether test files should be analyzed, too (default true)
		-trace
		  	trace output

# Examples

To check the current package:

	errortype .

To check errors across packages, using and suggesting overrides:

	errortype  -overrides=overrides.yaml -suggest=- ./...
*/
package main
