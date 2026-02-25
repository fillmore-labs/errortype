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

package analyze

import (
	"context"
	"runtime/trace"
	"slices"

	"fillmore-labs.com/errortype/detect/result"
	"fillmore-labs.com/errortype/internal/overrides"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// Suggestions generates a map of type names based on their determined usage.
func (p Pass) Suggestions(ctx context.Context) overrides.Overrides {
	defer trace.StartRegion(ctx, "suggestions").End()

	suggestions := overrides.Overrides{
		Types: make(map[result.ErrorType][]typeutil.TypeName),
	}

	for tn, usage := range p.ErrorUsage {
		typ := usage.DeterminedType()
		if typ == result.Undecided {
			continue
		}

		suggestions.Types[typ] = append(suggestions.Types[typ], typeutil.NewTypeName(tn))
	}

	for _, s := range suggestions.Types {
		slices.SortFunc(s, typeutil.TypeName.Compare)
	}

	return suggestions
}
