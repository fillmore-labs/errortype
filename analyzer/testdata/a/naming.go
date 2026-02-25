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

import "errors"

type ErrNamed string // want ` \(et:nam\+\)$`

func (e ErrNamed) Error() string { return string(e) }

const (
	sentinelError ErrNamed = "sentinelError" // want ` \(et:nam\)$`

	otherSentinel = ErrNamed("otherSentinel") // want ` \(et:nam\)$`
)

type RetryableError interface {
	error
	Retryable() bool
}

type myErrors struct{ errs []error }

func (e myErrors) Error() string {
	err := errors.Join(e.errs...)
	if err == nil {
		return "<nil>"
	}

	return err.Error()
}

func (e myErrors) Unwrap() []error {
	return e.errs
}
