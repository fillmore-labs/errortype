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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/checker"
	"golang.org/x/tools/go/analysis/unitchecker"
	"golang.org/x/tools/go/packages"

	"fillmore-labs.com/errortype/internal/analyze"
	"fillmore-labs.com/errortype/internal/detect"
	"fillmore-labs.com/errortype/internal/errortypes"
	"fillmore-labs.com/errortype/internal/overrides"
	"fillmore-labs.com/errortype/internal/typeutil"
)

func main() {
	a := analyze.Analyzer
	d := detect.Analyzer
	analyzers := []*analysis.Analyzer{a, d}

	log.SetFlags(0)
	log.SetPrefix(a.Name + ": ")

	if err := analysis.Validate(analyzers); err != nil {
		log.Fatal(err)
	}

	flags := setFlags(analyzers)

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if len(args) == 1 && strings.HasSuffix(args[0], ".cfg") {
		unitchecker.Run(args[0], analyzers)
		panic("unreachable")
	}

	cfg := &packages.Config{
		Mode:  packages.LoadAllSyntax,
		Tests: flags.IncludeTests,
	}

	pkgs, err := packages.Load(cfg, args...)
	if err != nil {
		log.Fatal(err) // failure to enumerate packages
	}

	if len(pkgs) == 0 {
		log.Fatalf("%s matched no packages", strings.Join(args, " "))
	}

	var exitErr int

	if n := packages.PrintErrors(pkgs); n > 0 {
		exitErr = 1
	}

	opts := &checker.Options{}

	graph, err := checker.Analyze(analyzers, pkgs, opts)
	if err != nil {
		log.Fatal(err)
	}

	// Don't print the diagnostics
	// but apply all fixes from the root actions.
	if flags.Fix {
		if err := applyFixes(graph.Roots, flags.Diff); err != nil {
			// Fail when applying fixes failed.
			log.Fatal(err)
		}
		// Don't proceed to print text/JSON,
		// and don't report an error
		// just because there were diagnostics.
		return
	}

	if flags.JSON {
		if err := graph.PrintJSON(os.Stdout); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := graph.PrintText(os.Stderr, flags.Context); err != nil {
			log.Fatal(err)
		}
	}

	if exitErr > 0 {
		os.Exit(exitErr)
	}

	if err := writeSuggestions(flags.Suggest, graph); err != nil {
		log.Fatal(err)
	}
}

func writeSuggestions(name string, graph *checker.Graph) error {
	if name == "" {
		return nil
	}

	suggestions := calculateSuggestions(graph)
	if len(suggestions) == 0 {
		return nil
	}

	var out *os.File
	if name == "-" {
		out = os.Stdout
	} else if suggest, err := os.OpenFile(filepath.Clean(name), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o666); err == nil { //nolint:gosec
		defer suggest.Close()
		out = suggest
	} else {
		return fmt.Errorf("can't write suggestion file: %w", err)
	}

	return overrides.Write(out, suggestions)
}

func calculateSuggestions(graph *checker.Graph) []overrides.Override {
	type ErrorType uint8

	const (
		Undecided ErrorType = 1 << iota
		PointerType
		ValueType
	)

	combined := make(map[typeutil.TypeName]ErrorType)

	for _, root := range graph.Roots {
		if r, ok := root.Result.(analyze.Result); ok {
			for _, p := range r.Pointers {
				combined[p] |= PointerType
			}

			for _, v := range r.Values {
				combined[v] |= ValueType
			}

			for _, v := range r.Inconsistent {
				combined[v] |= Undecided
			}
		}
	}

	if len(combined) == 0 {
		return nil
	}

	suggestions := make([]overrides.Override, 0, len(combined))

	for name, usage := range combined {
		switch usage {
		case PointerType:
			suggestions = append(suggestions, overrides.Override{TypeName: name, ErrorType: errortypes.PointerType})

		case ValueType:
			suggestions = append(suggestions, overrides.Override{TypeName: name, ErrorType: errortypes.ValueType})

		default:
			suggestions = append(suggestions, overrides.Override{TypeName: name, ErrorType: errortypes.Undecided})
		}
	}

	return suggestions
}
