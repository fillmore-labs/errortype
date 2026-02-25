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

// Package analyzer provides the errortype [golang.org/x/tools/go/analysis.Analyzer],
// which helps prevent subtle bugs in error handling. Use [Analyzer] for the default
// configuration, or [New] to configure it with [Option] values.
//
// See the [errortype command] for a description of the checks.
//
// [errortype command]: https://pkg.go.dev/fillmore-labs.com/errortype
package analyzer

import _ "embed"

// doc is the analyzer documentation, shown by -help. It is also the documentation
// of the errortype command; run "go generate" after editing doc.txt.
//
//go:generate go tool gendoc -name errortype
//go:embed doc.txt
var doc string
