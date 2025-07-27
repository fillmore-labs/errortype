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

package a

import (
	"errors"
	"fmt"
	"go/parser"
	"go/scanner"
	"net"
	"syscall"
)

func CheckError(err error) {
	switch e := err.(type) {
	case *net.UnknownNetworkError: // want " \\(et:ast\\)$"
		fmt.Println("UnknownNetworkError:", (string)(*e))

	case *net.InvalidAddrError: // want " \\(et:ast\\)$"
		fmt.Println("InvalidAddrError:", (string)(*e))

	case *syscall.Errno: // want " \\(et:ast\\)$"
		fmt.Println("Errno:", (int)(*e))
	}

	var e net.AddrError
	if errors.As(err, &e) { // want " \\(et:arg\\)$"
		fmt.Println("AddrError:", e.Err)
	}
}

func CheckError2() {
	_, err := parser.ParseFile(nil, "", nil, parser.AllErrors)
	if list, ok := err.(*scanner.ErrorList); ok { // want " \\(et:ast\\)$"
		fmt.Println(list)
	}
}
