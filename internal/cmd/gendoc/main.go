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

package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"

	"fillmore-labs.com/errortype/internal/gendoc"
	"fillmore-labs.com/errortype/internal/naming"
	"fillmore-labs.com/errortype/internal/settings"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("gendoc: ")

	name := flag.String("name", "", "analyzer `name`")

	flag.Parse()

	if err := run(*name); err != nil {
		log.Fatal(err)
	}
}

var (
	errMissingFlag     = errors.New("missing flag")
	errUnknownAnalyzer = errors.New("unknown analyzer")
)

func run(name string) error {
	if name == "" {
		return fmt.Errorf("%w: -name is required", errMissingFlag)
	}

	flags, err := analyzerFlags(name)
	if err != nil {
		return err
	}

	doc, err := os.ReadFile(gendoc.SourceFile)
	if err != nil {
		return err
	}

	content, err := gendoc.Generate(name, string(doc), flags)
	if err != nil {
		return fmt.Errorf("%s: %w", gendoc.SourceFile, err)
	}

	// Don't overwrite files that belong to something else.
	if _, err := gendoc.ReadCommandFile(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	return os.WriteFile(gendoc.CommandFile, content, 0o644) //nolint:gosec
}

// analyzerFlags returns the flags registered by the named analyzer.
//
// The flags come from the internal registration, not the analyzer packages,
// so generation does not depend on the generated doc constant.
func analyzerFlags(name string) (*flag.FlagSet, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)

	switch name {
	case "errortype":
		settings.DefaultConfig().RegisterFlags(fs)

	case "errorname":
		naming.DefaultOptions().RegisterFlags(fs)

	default:
		return nil, fmt.Errorf("%w %q", errUnknownAnalyzer, name)
	}

	return fs, nil
}
