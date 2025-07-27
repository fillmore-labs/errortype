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
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// analyzerFlags are the command line flags for the analyzer.
type analyzerFlags struct {
	// IncludeTests indicates whether test files should be analyzed too.
	IncludeTests bool

	// -json
	JSON bool

	// -c=N: if N>0, display offending line plus N lines of context
	Context int

	// Fix determines whether to apply (!Diff) or display (Diff) all suggested fixes.
	Fix bool

	// Diff causes the file updates to be displayed, but not applied.
	// This flag has no effect unless Fix is true.
	Diff bool

	// Suggest writes a file with suggestions.
	Suggest string
}

// defaultFlags are the default setting for the command line flags.
func defaultFlags() *analyzerFlags {
	return &analyzerFlags{
		IncludeTests: true,
		JSON:         false,
		Context:      -1,
		Fix:          false,
		Diff:         false,
	}
}

func setFlags(analyzers []*analysis.Analyzer) *analyzerFlags {
	for _, a := range analyzers {
		a.Flags.VisitAll(func(f *flag.Flag) {
			flag.Var(f.Value, f.Name, f.Usage)
		})
	}

	a := analyzers[0]

	flag.Usage = func() {
		paras := strings.Split(a.Doc, "\n\n")
		_, _ = fmt.Fprintf(os.Stderr, "%s: %s\n\n", a.Name, paras[0])
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s [-flag] [package]\n\n", a.Name)

		if len(paras) > 1 {
			_, _ = fmt.Fprintln(os.Stderr, strings.Join(paras[1:], "\n\n"))
		}

		_, _ = fmt.Fprintln(os.Stderr, "\nFlags:")

		flag.PrintDefaults()
	}

	f := defaultFlags()

	flag.BoolFunc("V", "print version and exit", version)
	flag.BoolFunc("flags", "print analyzer flags in JSON", printFlags)
	flag.BoolVar(&f.IncludeTests, "test", f.IncludeTests, "indicates whether test files should be analyzed, too")
	flag.BoolVar(&f.JSON, "json", f.JSON, "emit JSON output")
	flag.IntVar(&f.Context, "c", f.Context, `display offending line with this many lines of context`)
	// flag.BoolVar(&f.Fix, "fix", f.Fix, "apply all suggested fixes")
	// flag.BoolVar(&f.Diff, "diff", f.Diff, "with -fix, don't update the files, but print a unified diff")
	flag.StringVar(&f.Suggest, "suggest", f.Suggest, "append override suggestions to this file, - for standard output")

	return f
}

// version represents a [flag] to print version information and exit the program.
func version(string) error {
	progname, err := os.Executable()
	if err != nil {
		return err
	}

	if bi, ok := debug.ReadBuildInfo(); ok {
		fmt.Printf("%s version %s build with %s\n",
			filepath.Base(progname), bi.Main.Version, bi.GoVersion)
	} else {
		fmt.Printf("%s version (unknown)\n", filepath.Base(progname))
	}

	os.Exit(0)

	return nil
}

func printFlags(string) error {
	type jsonFlag struct {
		Name  string
		Bool  bool
		Usage string
	}

	var flags []jsonFlag

	flag.VisitAll(func(f *flag.Flag) {
		// Don't report flags that have no effect on unitchecker
		// (as invoked by 'go vet').
		switch f.Name {
		case "fix", "diff":
			return
		}

		var isBool bool
		if b, ok := f.Value.(interface{ IsBoolFlag() bool }); ok {
			isBool = b.IsBoolFlag()
		}

		flags = append(flags, jsonFlag{f.Name, isBool, f.Usage})
	})

	data, err := json.MarshalIndent(flags, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	_, _ = os.Stdout.Write(data)
	os.Exit(0)

	return nil
}
