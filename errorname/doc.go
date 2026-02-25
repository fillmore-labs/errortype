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

/*
errorname is a static analyzer that checks that error variables use an "Err" prefix
(e.g., ErrNotFound) and structured error types use an "Error" suffix (e.g., ParseError).
This check is also available as part of [errortype].

Usage:

	errorname [flags] [packages]

The flags are:

	-c int
		display offending line with this many lines of context (default -1)
	-fix
		apply all suggested fixes
	-test
		indicates whether test files should be analyzed, too (default true)
	-generated
		indicates whether generated files should be analyzed (default false)

# Examples

To check the current package:

	errorname .

To fix error naming across packages:

	go fix -fixtool=$(which errorname) ./...

[errortype]: https://pkg.go.dev/fillmore-labs.com/errortype#section-readme
*/
package main
