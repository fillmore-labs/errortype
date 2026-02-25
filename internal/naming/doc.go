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

// Package naming implements the errorname analyzer, which renames error
// variables and types to follow Go's [error naming] convention: error variables
// begin with "Err" or "err", and error types end in "Error".
//
// The analyzer offers automated fixes. In addition to renaming the
// declarations and their usages, it automatically updates doc comments and,
// for exported package-level declarations, safely preserves backward
// compatibility by inserting "Deprecated:" aliases and "//go:fix inline"
// directives.
//
// [error naming]: https://go.dev/wiki/Errors#naming
package naming
