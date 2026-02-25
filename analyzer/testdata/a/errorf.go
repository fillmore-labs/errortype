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

package a

import (
	"fmt"

	"test/a/b"
	"test/a/d"
)

func _() {
	_ = fmt.Errorf("%w", &d.ValueError{}) // want ` \(et:wrp\)$`

	_ = fmt.Errorf("%w", d.PointerError{}) // want ` \(et:wrp\+\)$`

	_ = fmt.Errorf("%w", &b.AmbiguousError{}) // want ` \(et:emb\+\)$`

	// "%%" is a literal percent sign and does not consume an operand.
	_ = fmt.Errorf("100%% %w", &d.ValueError{}) // want ` \(et:wrp\)$`

	_ = fmt.Errorf("%%w", &d.ValueError{})

	// Flags, width and precision do not affect the operand mapping, ...
	_ = fmt.Errorf("%+w", &d.ValueError{})    // want ` \(et:wrp\)$`
	_ = fmt.Errorf("%4.2w", d.PointerError{}) // want ` \(et:wrp\+\)$`

	// ... but a width or precision given as "*" consumes an operand.
	_ = fmt.Errorf("%*w", 4, &d.ValueError{})            // want ` \(et:wrp\)$`
	_ = fmt.Errorf("%.*s %w", 4, "pad", &d.ValueError{}) // want ` \(et:wrp\)$`

	// Explicit argument indexes reposition the operands.
	_ = fmt.Errorf("%[2]w", "arg", &d.ValueError{})       // want ` \(et:wrp\)$`
	_ = fmt.Errorf("%[2]s %[1]w", &d.ValueError{}, "arg") // want ` \(et:wrp\)$`

	// Multiple %w verbs wrap multiple operands.
	_ = fmt.Errorf("%w %w", &d.ValueError{}, d.PointerError{}) // want ` \(et:wrp\)$` ` \(et:wrp\+\)$`

	// Operands of verbs other than %w are not wrapped.
	_ = fmt.Errorf("%v", &d.ValueError{})
	_ = fmt.Errorf("%s %d", &d.ValueError{}, 42)

	// A %w verb without a usable operand wraps nothing.
	_ = fmt.Errorf("%v: %w", &d.ValueError{})
	_ = fmt.Errorf("%[3]w", &d.ValueError{})

	// A non-constant format string cannot be analyzed.
	format := "%w"
	_ = fmt.Errorf(format, &d.ValueError{})

	// Spread operands cannot be matched to verbs.
	args := []any{&d.ValueError{}}
	_ = fmt.Errorf("%w", args...)
}
