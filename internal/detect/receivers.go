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

package detect

import (
	"go/types"

	"fillmore-labs.com/errortype/internal/errortypes"
	"fillmore-labs.com/errortype/internal/typeutil"
)

func (p pass) processReceivers() {
	for tn, errorType := range p.PropertyMap {
		if errorType.DeterminedType() != errortypes.Undecided {
			continue
		}

		if tn.Pkg() != p.Pkg {
			continue
		}

		named, ok := tn.Type().(*types.Named)
		if !ok {
			continue
		}

		ptr, pure := pureReceivers(named)
		if !pure {
			continue
		}

		property := ValueReceivers
		if ptr {
			property = PointerReceivers
		}

		p.AddTypeProperty(tn, property)
	}
}

// pureReceivers checks whether all methods of the given named type have receivers of the same kind
// (either all pointer receivers or all value receivers).
//
// It returns ok when the type has at least one method and all methods have receivers of the same kind,
// ptr when all methods have pointer receivers.
func pureReceivers(named *types.Named) (ptr, ok bool) {
	numMethods := named.NumMethods()
	if numMethods == 0 {
		return false, false
	}

	_, ptr0 := typeutil.HasPointerReceiver(named.Method(0).Signature())
	for i := 1; i < numMethods; i++ {
		if _, ptr1 := typeutil.HasPointerReceiver(named.Method(i).Signature()); ptr1 != ptr0 {
			return false, false
		}
	}

	return ptr0, true
}
