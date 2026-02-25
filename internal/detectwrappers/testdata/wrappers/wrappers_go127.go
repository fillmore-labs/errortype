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

//go:build go1.27

package wrappers

import "errors"

type GSuite[S any] struct{}

func (s GSuite[S]) Is[E error](err, target error) bool { // want Is:`is\(0, 1\)`
	return errors.Is(err, target)
}

func (s GSuite[A]) AsType[E error](err error) (E, bool) { // want AsType:`astype\(0, 0\)`
	return errors.AsType[E](err)
}
