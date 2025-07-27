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

package typeutil

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

var errGOROOTMismatch = errors.New("'go env GOROOT' does not match runtime.GOROOT")

// HasGo is a tool modeled after NeedsTool from "golang.org/x/tools/internal/testenv".
//
// See also https://github.com/bazel-contrib/rules_go/issues/3934
//
//nolint:staticcheck
func HasGo() error {
	_, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("go not found: %w", err)
	}

	if runtime.GOROOT() == "" {
		return nil
	}

	out, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		exit := &exec.ExitError{}
		if errors.As(err, &exit) {
			err = fmt.Errorf("%w\nstderr:\n%s)", err, exit.Stderr)
		}

		return err
	}

	if GOROOT := strings.TrimSpace(string(out)); GOROOT != runtime.GOROOT() {
		err := fmt.Errorf("%w:\n\tgo env: %s\n\tGOROOT: %s", errGOROOTMismatch, GOROOT, runtime.GOROOT())

		return err
	}

	return nil
}
