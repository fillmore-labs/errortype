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

// EmbErr is an exported type embedded by name is renamed with a deprecation alias;
// the embedding site and field selectors keep the old name via the alias.
type EmbErr struct{ msg string } // want ` suggestion: "EmbError"$`

func (e *EmbErr) Error() string { return e.msg }

type ErrEmbContainer struct { // want ` suggestion: "EmbContainerError"$`
	*EmbErr
}

func _() error {
	err := ErrEmbContainer{EmbErr: &EmbErr{msg: "boom"}}
	_ = err.EmbErr

	return err
}

func _() error {
	err := ErrEmbContainer{&EmbErr{msg: "boom"}}
	_ = err.EmbErr

	return err
}

// errEmb is unexported, but still embeddded.
type errEmb struct{ msg string } // want ` suggestion: "embError"$`

func (e *errEmb) Error() string { return e.msg }

func _() error {
	type emb2ContainerError struct {
		errEmb
	}

	err := &emb2ContainerError{errEmb: errEmb{msg: "boom"}}
	_ = err.errEmb

	return err
}

// errEmb2 is unexported, but still embeddded.
type errEmb2 struct{ msg string } // want ` suggestion: "emb2Error"$`

func (e errEmb2) Error() string { return e.msg }

func _() error {
	err := &struct {
		*errEmb2
	}{errEmb2: &errEmb2{msg: "boom"}}

	_ = err.errEmb2

	return err
}
