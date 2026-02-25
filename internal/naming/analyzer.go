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

import (
	"flag"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Documentation constants.
const (
	name = "errorname"
	url  = "https://pkg.go.dev/fillmore-labs.com/errortype/internal/cmd/errorname"
)

// Analyzer returns a new [analysis.Analyzer] configured for error naming checking,
// documented by doc.
func (o *Options) Analyzer(doc string) *analysis.Analyzer {
	a := &analysis.Analyzer{
		Name: name,
		Doc:  strings.TrimSpace(doc),
		URL:  url,
		Run:  o.Run,
	}

	o.RegisterFlags(&a.Flags)

	return a
}

// RegisterFlags binds the [Options] values to command line flag values.
func (o *Options) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&o.Generated, "generated", o.Generated, "include generated files")
}
