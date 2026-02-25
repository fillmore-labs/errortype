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

package typeutil

import "go/types"

// SignatureOf returns the signature of a func or function-typed variable.
func SignatureOf(obj types.Object) *types.Signature {
	switch o := obj.(type) {
	case *types.Func:
		return o.Signature()

	case *types.Var:
		sig, _ := o.Type().Underlying().(*types.Signature)
		return sig

	default:
		return nil
	}
}
