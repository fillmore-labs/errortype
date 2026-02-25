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

/*
Command gendoc generates the documentation of an analyzer command from the
analyzer documentation, see [fillmore-labs.com/errortype/internal/gendoc].

It reads doc.txt in the current directory, the analyzer package, and writes
doc.go of the command in the parent directory. A doc.go that was not generated
is not overwritten.

Usage:

	gendoc -name NAME

Intended to be run via go generate from the analyzer package, next to the
embedded documentation:

	//go:generate go tool gendoc -name errortype

	//go:embed doc.txt
	var doc string

Check that the generated file is up to date with
[fillmore-labs.com/errortype/internal/gendoc/gendoctest.CheckUpToDate].
*/
package main
