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

package usage

import (
	"context"
	"runtime/trace"

	"fillmore-labs.com/errortype/internal/overrides"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// Suggestions generates a map of type names based on their determined usage.
func (e ErrorUsage) Suggestions(ctx context.Context) overrides.Overrides {
	defer trace.StartRegion(ctx, "suggestions").End()

	suggestions := make(overrides.Overrides)

	for tn, typ := range e.allDetermined {
		suggestions[typ] = append(suggestions[typ], typeutil.NewTypeName(tn))
	}

	return suggestions
}
