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
	"golang.org/x/tools/go/analysis"

	"fillmore-labs.com/errortype/internal/errortypes"
)

// pass wraps an *analysis.pass and tracks error usages within the analysis pass.
// It contains a PropertyMap that associates error usages with their corresponding Usage information.
type pass struct {
	*analysis.Pass
	errorUsages errortypes.PropertyMap[Usage]
}

// newPass creates and returns a new Pass instance, initializing its errorUsages property map.
// It takes an *analysis.Pass as input and embeds it within the returned Pass.
func newPass(ap *analysis.Pass) pass {
	return pass{
		Pass:        ap,
		errorUsages: errortypes.NewPropertyMap[Usage](),
	}
}
