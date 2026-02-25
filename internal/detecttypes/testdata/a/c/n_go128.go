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

//go:build go1.28

package c

// Composite literals whose (pointer) type is inferred from
// the assignment context (Go 1.28, go.dev/issue/82039).

type (
	InferredSliceElemPtrError struct{ error } // want InferredSliceElemPtrError:"pointer"
	InferredSliceElemValError struct{ error } // want InferredSliceElemValError:"value"
	InferredLitPtrError       struct{ error } // want InferredLitPtrError:"pointer"
	InferredLitValError       struct{ error } // want InferredLitValError:"value"
	InferredPtrSliceError     struct{ error } // want InferredPtrSliceError:"pointer"
)

func testInferredLiterals(p *InferredSliceElemPtrError, v InferredSliceElemValError) {
	var _ *[]error = {p}
	var _ *[]error = {v}
	var _ *InferredLitPtrError = {}
	var _ InferredLitValError = {}
	var _ *[]*InferredPtrSliceError = {}
}
