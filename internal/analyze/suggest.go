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

package analyze

import (
	"fmt"
	"os"
	"path/filepath"

	"fillmore-labs.com/errortype/internal/errortypes"
	"fillmore-labs.com/errortype/internal/overrides"
	"fillmore-labs.com/errortype/internal/typeutil"
)

func (p pass) calculateSuggestions() map[errortypes.ErrorType][]typeutil.TypeName {
	suggestions := make(map[errortypes.ErrorType][]typeutil.TypeName)

	for tn, typ := range p.errorUsages.AllDetermined {
		if typ == errortypes.SuppressType {
			continue
		}

		suggestions[typ] = append(suggestions[typ], typeutil.NewTypeName(tn))
	}

	return suggestions
}

func (o *Options) writeSuggestions(suggestions map[errortypes.ErrorType][]typeutil.TypeName, name string) error {
	if o.Suggest == "" || len(suggestions) == 0 {
		return nil
	}

	var out *os.File
	if o.Suggest == "-" {
		out = os.Stdout
	} else {
		o.suggestwrite.Lock()
		defer o.suggestwrite.Unlock()

		suggest, err := os.OpenFile(filepath.Clean(o.Suggest), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o666) //nolint:gosec
		if err != nil {
			return fmt.Errorf("can't write suggestion file: %w", err)
		}

		defer suggest.Close()
		out = suggest
	}

	return overrides.Write(out, suggestions, name)
}
