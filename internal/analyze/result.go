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
	"fillmore-labs.com/errortype/internal/errortypes"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// Result holds the categorized type names found during analysis.
type Result struct {
	// Pointers contains type names consistently (but maybe wrongly) used as pointer types.
	Pointers []typeutil.TypeName

	// Values contains type names consistently (but maybe wrongly) used as value types.
	Values []typeutil.TypeName

	// Inconsistent contains type names that were found to be used inconsistently as both pointer and value types.
	Inconsistent []typeutil.TypeName
}

// calculateResult analyzes the collected error usages and categorizes them into pointers,
// values, and inconsistent types.
func (p pass) calculateResult() Result {
	var pointers, values, inconsistent []typeutil.TypeName
	for tn, typ := range p.errorUsages.AllDetermined {
		typeName := typeutil.NewTypeName(tn)

		switch typ {
		case errortypes.PointerType:
			pointers = append(pointers, typeName)

		case errortypes.ValueType:
			values = append(values, typeName)

		case errortypes.SuppressType:
			inconsistent = append(inconsistent, typeName)
		}
	}

	return Result{Pointers: pointers, Values: values, Inconsistent: inconsistent}
}
