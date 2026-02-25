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
	"log/slog"

	"fillmore-labs.com/errortype/internal/detectwrappers/result"
	"fillmore-labs.com/errortype/internal/settings"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// WithWrappers returns an Option that applies the provided overrides mapping,
// allowing specific type names to be associated with custom error types.
func WithWrappers(overrides map[result.WrapperType][]typeutil.FuncName) Option {
	return wrappersOption{overrides: overrides}
}

type wrappersOption struct {
	overrides map[result.WrapperType][]typeutil.FuncName
}

func (o wrappersOption) Apply(opts *settings.Config) error {
	opts.WrapperConfig.AddWrappers(o.overrides)
	return nil
}

func (o wrappersOption) LogAttr() slog.Attr {
	return logAttrMap("overridesw", o.overrides)
}
