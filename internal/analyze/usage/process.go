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

	"fillmore-labs.com/errortype/facts"
)

// ProcessDetectedTypes populates the initial error usage map based on the results
// from the prerequisite `detecttypes` analyzer.
func (e ErrorUsage) ProcessDetectedTypes(ctx context.Context, resultInfo []facts.ResultInfo) {
	defer trace.StartRegion(ctx, "detectedTypes").End()

	for _, detectedType := range resultInfo {
		var usage Usage

		switch detectedType.ErrorType & facts.ExpectedMask {
		case facts.PointerType:
			usage = PointerExpected

		case facts.ValueType:
			usage = ValueExpected

		case facts.SuppressType:
			usage = SuppressExpected

		default:
			continue
		}

		e[detectedType.TypeName] = usage
	}
}
