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

package naming

import "strconv"

type ErrNamed string // want ` \(et:nam\+\)$`

func (e ErrNamed) Error() string { return string(e) }

const (
	sentinelError ErrNamed = "sentinelError" // want ` \(et:nam\)$`

	otherSentinel = ErrNamed("otherSentinel") // want ` \(et:nam\)$`
)

type RetryableErr interface { // want ` \(et:nam\+\)$`
	error
	Retryable() bool
}

type NumErr int // want ` \(et:nam\+\)$`

func (e NumErr) Error() string {
	return "error " + strconv.Itoa(int(e))
}

const (
	NumErr0 NumErr = 0    // want ` \(et:nam\)$`
	NumErr1 NumErr = iota // want ` \(et:nam\)$`
	NumErr2               // want ` \(et:nam\)$`
)

const (
	NumErr3 NumErr = 3         // want ` \(et:nam\)$`
	NumErr4 NumErr = 4         // want ` \(et:nam\)$`
	NumErr5        = NumErr(5) // want ` \(et:nam\)$`
)
