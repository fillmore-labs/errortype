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

// Package analyze provides the main errortype analyzer for detecting and enforcing
// consistent error type usage in Go programs.
//
// The analyzer identifies several types of error handling issues:
//
//   - Inconsistent pointer/value usage in function returns
//   - Incorrect type assertions on error types
//   - Misuse of errors.As with wrong pointer/value semantics
//   - Switch statements with inconsistent error type handling
//
// This package integrates with the golang.org/x/tools/go/analysis framework
// and depends on the internal/detect package for error type discovery.
package analyze
