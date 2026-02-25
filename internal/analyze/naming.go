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

package analyze

import (
	"go/types"

	"fillmore-labs.com/errortype/internal/naming/diagnostic"
)

// namingMessage annotates the default error-naming message with the
// matching errorname code.
func namingMessage(obj types.Object, newName string) string {
	msg := diagnostic.DefaultMessage(obj, newName)
	if msg == "" {
		return ""
	}

	switch obj.(type) {
	case *types.TypeName:
		return msg + " (et:nam+)"

	default:
		return msg + " (et:nam)"
	}
}
