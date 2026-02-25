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

package testdata

import "errors"

var (
	errorPrefix  = errors.New("error prefix")  // want ` suggestion: "errPrefix"$`
	ErrorPrefix  = errors.New("error prefix")  // want ` suggestion: "ErrPrefix"$`
	errorsPrefix = errors.New("errors prefix") // want ` suggestion: "errPrefix2"$`
	ErrorsPrefix = errors.New("errors prefix") // want ` suggestion: "ErrPrefix2"$`
	errlower     = errors.New("err lower")     // want ` suggestion: "errLower"$`
	Errlower     = errors.New("err lower")     // want ` suggestion: "ErrLower"$`
)

const (
	ConstOmitted1Err NumErr = 10 // want ` suggestion: "ErrConstOmitted1"$`
	ConstOmitted2Err             // want ` suggestion: "ErrConstOmitted2"$`
)
