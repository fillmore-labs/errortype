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

package analyzer

import (
	"fmt"
	"log/slog"

	"fillmore-labs.com/errortype/internal/detectwrappers/result"
	"fillmore-labs.com/errortype/internal/qualname"
	"fillmore-labs.com/errortype/internal/settings"
)

// WithWrappers returns an [Option] that declares functions as wrappers of error functions,
// for wrappers that are not detected automatically.
//
// The keys name the wrapped function: "is" ([errors.Is]), "as" ([errors.As]),
// "astype" ([errors.AsType]) or "errorf" ([fmt.Errorf]). The values are fully qualified
// function or method names:
//
//	analyzer.WithWrappers(map[string][]string{
//		"is": {"example.com/pkg.IsError", "(example.com/pkg.Checker).Is"},
//	})
//
// Invalid keys or function names are reported by [New].
func WithWrappers(wrappers map[string][]string) Option {
	return wrappersOption{wrappers: wrappers}
}

type wrappersOption struct {
	wrappers map[string][]string
}

func (o wrappersOption) Apply(opts *settings.Config) error {
	wrappers, err := parseNames[result.WrapperType, qualname.Func](o.wrappers, parseWrapperType)
	if err != nil {
		return fmt.Errorf("wrappers: %w", err)
	}

	opts.WrapperConfig.AddWrappers(wrappers)

	return nil
}

func parseWrapperType(kind string) (result.WrapperType, error) {
	var w result.WrapperType
	err := w.UnmarshalText([]byte(kind))

	return w, err
}

func (o wrappersOption) LogAttr() slog.Attr {
	return logAttrMap("wrappers", o.wrappers)
}
