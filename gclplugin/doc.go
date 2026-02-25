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
Package gclplugin provides golangci-lint plugin integration for the [errortype] analyzer.

# Usage

1. Add a file `.custom-gcl.yaml` to your source with:

	---
	version: v2.15.0

	name: golangci-lint
	destination: .

	plugins:
	  - module: fillmore-labs.com/errortype
	    import: fillmore-labs.com/errortype/gclplugin
	    version: v0.0.13

2. Run `golangci-lint custom` from your project root.

This will create a custom `golangci-lint` executable in your project root.

3. Configure the linter in `.golangci.yaml`:

	---
	version: '2'
	linters:
	  default: none
	  enable:
	    - errortype
	  settings:
	    custom:
	      errortype:
	        type: module
	        description: checks pointer/value usage, comparisons and naming of Go error types.
	        original-url: https://fillmore-labs.com/errortype
	        settings:
	          # All recommended checks, for local development:
	          recommended: true
	          # In CI, enable the individual checks your codebase passes instead:
	          # naming: true
	          # legacy: true
	          # receivers: true
	          # non-comparable: true
	          # style-check: true
	          strict-is-target: false
	          strict-is-method: false
	          unchecked-assert: false
	          check-unused: true
	          prefix-filter: true
	          overrides:
	            pointer:
	              - your.pkg/a.PointerOverrideError
	            value:
	              - your.pkg/a.ValueOverrideError
	            suppress:
	              - your.pkg/a.SuppressOverrideError

4. Run the linter:

	./golangci-lint run .

[errortype]: https://pkg.go.dev/fillmore-labs.com/errortype#section-readme
*/
package gclplugin
