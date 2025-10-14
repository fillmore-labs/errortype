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

// Package knownfuncs provides metadata about known error-handling and assertion functions.
//
// This package maintains a registry of functions that behave like errors.Is, errors.As,
// and testing assertion functions from popular libraries. It classifies these functions
// by their behavior patterns and provides structured information about their signatures
// and parameter positions.
//
// The package supports functions from:
//   - Standard library (errors, reflect)
//   - Error handling libraries (golang.org/x/xerrors, github.com/pkg/errors, etc.)
//   - Testing frameworks (github.com/stretchr/testify, gotest.tools)
//   - Custom error handling libraries (fillmore-labs.com/exp/errors)
//
// Function Classifications:
//
// KindIs - Functions that compare errors (like errors.Is):
//   - IsFunc0: No additional context parameter (e.g., errors.Is(err, target))
//   - IsFunc1: One additional context parameter (e.g., assert.ErrorIs(t, err, target))
//
// KindAs - Functions that extract error types (like errors.As):
//   - AsFunc0: No additional context parameter (e.g., errors.As(err, &target))
//   - AsFunc1: One additional context parameter (e.g., assert.ErrorAs(t, err, &target))
//   - AssertFunc: Type assertion without target pointer (e.g., errors.AsType[MyError](err))
//
// KindEqu - Functions that compare equality (like assert.Equal):
//   - Similar parameter patterns to KindIs functions
//
// KindType - Functions that check type identity (like assert.IsType):
//   - Similar parameter patterns to KindIs functions
//
// The FuncInfo struct encodes this metadata in a compact format, including:
//   - Function kind (Is/As/Equal/Type)
//   - Target argument index
//   - Type parameter index (for generic functions)
//   - Evaluation requirement (whether return value must be checked)
package knownfuncs
