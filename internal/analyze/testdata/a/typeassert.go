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
	"crypto/aes"
	"net"

	"golang.org/x/net/http2"
)

func Assert() {
	var err error

	switch err.(type) {
	case *aes.KeySizeError: // want " \\(et:ast\\)$"
	}

	switch e := err.(type) {
	case *http2.StreamError: // want " \\(et:ast\\)$"
		_ = e.StreamID
	}

	_, _ = err.(*aes.KeySizeError) // want " \\(et:ast\\)$"

	_, _ = err.(interface{ Temporary() bool })

	_, _ = err.(struct{ *net.ParseError })
}

func Assert2() {
	var err interface {
		Error() string
	}

	switch err.(type) {
	case net.InvalidAddrError:
	}

	switch e := err.(type) {
	case net.UnknownNetworkError:
		_ = e.Temporary()
	}

	_, _ = err.(*net.ParseError)

	_, _ = err.(interface{ Temporary() bool })

	_, _ = err.(struct{ *net.ParseError })
}

func Assert3() {
	var err net.Error

	switch err.(type) {
	case net.InvalidAddrError:
	case *net.InvalidAddrError: // want " \\(et:ast\\)$"
	default:
	}

	switch e := err.(type) {
	case net.UnknownNetworkError:
		_ = e.Temporary()
	case *net.UnknownNetworkError: // want " \\(et:ast\\)$"
		_ = e.Temporary()
	case nil:
	}

	_, _ = err.(*net.ParseError)

	_, _ = err.(interface{ Temporary() bool })

	_, _ = err.(struct{ *net.ParseError })
}
