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

package typeutil

import "go/types"

// anyType is the `any` interface type from the universal scope.
var anyType = types.Universe.Lookup("any").Type()

// IsAnyInterface determines if the given type is the `any` interface or an interface with no methods.
func IsAnyInterface(typ types.Type) bool {
	if typ == anyType {
		return true
	}

	iface, ok := typ.Underlying().(*types.Interface)

	return ok && iface.NumMethods() == 0
}
