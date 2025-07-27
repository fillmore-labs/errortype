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

// Package detect implements a multi-stage usage analysis for Go error types.
// Its primary goal is to determine whether a custom error type should be
// consistently used as a pointer (*T) or a value (T).
//
// The analysis employs a series of heuristics in a specific order of precedence:
//  1. The type has an Error() method with a pointer receiver. These error types
//     are only usable as pointer type s.
//  2. User-defined overrides from an external file, which can explicitly set an
//     error as a pointer type, a value type, or suppress analysis.
//  3. Variable declarations, such as compile-time assertions (`var _ error = T{}`)
//     or sentinel errors (`var ErrSomething = &T{}`).
//  4. An inspection of usage in function bodies, such as return statements, type
//     assertions, composite literals, and type casts.
//  5. As a final heuristic, a check for consistent receiver types (all pointer
//     or all value) across all methods of the error type.
//
// The determindes error types are passed as facts across packages and as a result
// for the errortype analyzer to use.
package detect
