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

package diagnostic

import (
	"fmt"
	"go/types"

	"fillmore-labs.com/errortype/internal/naming/rules"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// MessageFunc builds the diagnostic message for a rename. Returning the empty
// string suppresses the diagnostic for that object.
type MessageFunc func(obj types.Object, newName string) string

// DefaultMessage formats the default diagnostic message for renaming obj to
// newName. It returns the empty string for object kinds that are not renamed.
func DefaultMessage(obj types.Object, newName string) string {
	switch obj.(type) {
	case *types.Const, *types.Var:
		return fmt.Sprintf("Error %s %q should begin with %q, suggestion: %q", varKind(obj), obj.Name(), varPrefix(obj), newName)

	case *types.TypeName:
		return fmt.Sprintf("Error type %q should end with %q, suggestion: %q", obj.Name(), rules.TypeSuffix, newName)

	default:
		return ""
	}
}

// varKind returns "sentinel" for package-level objects or "variable" otherwise.
func varKind(obj types.Object) string {
	if typeutil.PackageLevel(obj) {
		return "sentinel"
	}

	return "variable"
}

// varPrefix returns the prefix for an error variable.
func varPrefix(obj types.Object) string {
	if obj.Exported() {
		return rules.VarPrefixExported
	}

	return rules.VarPrefix
}
