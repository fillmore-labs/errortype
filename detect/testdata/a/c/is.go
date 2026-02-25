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

package c

type (
	WithIs1Error    struct{ error } // want WithIs1Error:"pointer"
	WithIs2Error    struct{ error } // want WithIs2Error:"undecided"
	WithoutIs1Error struct{ error } // want WithoutIs1Error:"undecided"
	WithoutIs2Error struct{ error } // want WithoutIs2Error:"undecided"
)

func (e *WithIs1Error) Is(err error) bool {
	return e.error == err
}

func (e WithIs2Error) Is(err error) bool {
	return e.error == err
}

func (e *WithoutIs1Error) Is(err1, err2 error) bool {
	return e.error == err1
}

func (e *WithoutIs2Error) Is(error) string {
	return e.error.Error()
}

func (e WithIs1Error) Whatever()    {}
func (e *WithIs2Error) Whatever()   {}
func (e WithoutIs1Error) Whatever() {}
func (e WithoutIs2Error) Whatever() {}
