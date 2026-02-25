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

package run

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"fillmore-labs.com/errortype/internal/overrides"
)

func (o *Options) writeSuggestions(ctx context.Context, suggestions overrides.Overrides, pkgPath string) (err error) {
	if o.Suggest == "" || (len(suggestions.Types) == 0 && len(suggestions.Wrappers) == 0) {
		return nil
	}

	o.suggestwrite.Lock()
	defer o.suggestwrite.Unlock()

	var out *os.File
	if o.Suggest == "-" {
		out = os.Stdout
	} else {
		out, err = os.OpenFile(filepath.Clean(o.Suggest), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644) // #nosec G302 -- no sensitive data.
		if err != nil {
			return fmt.Errorf("can't write suggestion file: %w", err)
		}

		defer func() {
			if closeErr := out.Close(); closeErr != nil {
				err = errors.Join(err, closeErr)
			}
		}()
	}

	return overrides.Write(ctx, out, suggestions, pkgPath)
}
