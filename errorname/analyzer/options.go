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

	"fillmore-labs.com/errortype/internal/naming"
	"fillmore-labs.com/errortype/internal/options"
)

// Option configures specific behavior of the detect [analysis.Analyzer].
type Option interface {
	Apply(opts *naming.Options) error
	LogAttr() slog.Attr
}

// Join creates a new Option joining the provided Option values.
//
// The result implements [slog.LogValuer], so the following is evaluated lazily:
//
//	slog.LogAttrs(ctx, slog.LevelInfo, "settings", Join(opts...).LogAttr())
func Join(opts ...Option) Option {
	return options.Join(opts)
}

// WithGenerated is an [Option] to enable reports in generated files.
func WithGenerated(generated bool) Option {
	return generatedOption{generated: generated}
}

type generatedOption struct{ generated bool }

func (o generatedOption) Apply(opts *naming.Options) error {
	opts.Generated = o.generated

	return nil
}

func (o generatedOption) LogAttr() slog.Attr {
	return slog.Bool("generated", o.generated)
}
