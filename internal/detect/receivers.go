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
	"context"
	"go/types"
	"runtime/trace"

	"fillmore-labs.com/errortype/detect/result"
	"fillmore-labs.com/errortype/internal/detect/properties"
	"fillmore-labs.com/errortype/internal/typeutil"
)

func (p pass) processReceivers(ctx context.Context) {
	defer trace.StartRegion(ctx, "receivers").End()

	for tn, errorType := range p.ErrorTypes {
		if errorType.DeterminedType() != result.Undecided {
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

		property := properties.ValueReceivers
		if ptr {
			property = properties.PointerReceivers
		}

		p.ErrorTypes[tn] |= property
	}
}

// pureReceivers checks whether all methods of the given named type have receivers of the same kind
// (either all pointer receivers or all value receivers).
//
// It returns ok when the type has at least one method and all methods have receivers of the same kind,
// ptr when all methods have pointer receivers.
func pureReceivers(named *types.Named) (ptr, ok bool) {
	first := true

	for m := range named.Methods() {
		switch _, ptrcv := typeutil.IsPointerReceiver(m.Signature().Recv()); {
		case first:
			ptr, ok = ptrcv, true
			first = false

		case ptr != ptrcv:
			return false, false
		}
	}

	return ptr, ok
}
