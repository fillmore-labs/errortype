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

package detect

import (
	"flag"

	"fillmore-labs.com/errortype/internal/detect"
)

// registerFlags binds the [detect.Options] values to command line flag values.
// A nil flag set value defaults to the program's command line.
func registerFlags(o *detect.Options, fs *flag.FlagSet) {
	if fs == nil {
		fs = flag.CommandLine
	}

	fs.Func("overrides", "read error type overrides from this `file`", o.ReadOverrides)
	fs.Func("heuristics", "`list` of heuristics used (default: \"var,usage,receivers\", \"off\" to disable)", o.SetHeuristics)
	fs.Func("tracetypes", "information of error type detection in packages matching this `regex`", o.SetTrace)
}
