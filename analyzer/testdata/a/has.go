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

package a

import "fillmore-labs.com/exp/errors"

func Has() {
	var err error

	_, _ = errors.Has[*IntError](err)        // want " \\(et:ast\\)$"
	_, _ = errors.HasError[StringError](err) // want " \\(et:ast\\+\\)$"

	var itarget *IntError
	_ = errors.As(err, &itarget)            // want " \\(et:err\\)$"
	_ = errors.As[*IntError](err, &itarget) // want " \\(et:ast\\)$"

	var starget StringError
	_ = errors.As(err, &starget)              // want " \\(et:err\\+\\)$"
	_ = errors.As[StringError](err, &starget) // want " \\(et:ast\\+\\)$"
}
