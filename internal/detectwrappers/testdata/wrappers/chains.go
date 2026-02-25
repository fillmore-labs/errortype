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

package wrappers

import (
	"errors"
	"net"
)

func wrapIs(err, target error) bool { // want wrapIs:`is\(0, 1\)`
	return errors.Is(err, target)
}

func wrapIs2(err, target error) bool { // want wrapIs2:`is\(0, 1\)`
	return wrapIs(err, target)
}

func wrapIs3(err, target error) bool { // want wrapIs3:`is\(0, 1\)`
	return wrapIs2(err, target)
}

func wrapAs(err error, target any) bool { // want wrapAs:`as\(0, 1\)`
	return errors.As(err, target)
}

func wrapAs2(err error, target any) bool { // want wrapAs2:`as\(0, 1\)`
	return wrapAs(err, target)
}

func wrapAs3(err error, target any) bool { // want wrapAs3:`as\(0, 1\)`
	return wrapAs2(err, target)
}

func wrapAsType[E error](err error) (e E, ok bool) { // want wrapAsType:`astype\(0, 0\)`
	ok = errors.As(err, &e)
	return
}

func wrapAsType2[E any, F error](err2 error) (F, bool) { // want wrapAsType2:`astype\(0, 1\)`
	return wrapAsType[F](err2)
}

func wrapAsType3[F error, G any](err3 error) (F, bool) { // want wrapAsType3:`astype\(0, 0\)`
	return wrapAsType2[G, F](err3)
}

func xwrapAsType2[F error](err2 error, _ *F) (F, bool) { // want xwrapAsType2:`astype\(0, 0\)`
	return wrapAsType[F](err2)
}

func xwrapAsType3[G error](err3 error) (G, bool) { // want xwrapAsType3:`astype\(0, 0\)`
	return xwrapAsType2(err3, new(G))
}

func chainWrap(err error, target any) bool { // want chainWrap:`as\(0, 1\)`
	return errors.As(err, target)
}

func chainWrap2(err error, target any) bool { // want chainWrap2:`as\(0, 1\)`
	return chainWrap(err, target)
}

func chainWrap3[T error](err error) bool { // want chainWrap3:`astype\(0, 0\)`
	return chainWrap2(err, new(T))
}

func noWrapIs(err, target error) bool {
	return errors.As(err, &target)
}

func wrapAsNetType[E net.Error](err error) (e E, ok bool) { // want wrapAsNetType:`astype\(0, 0\)`
	ok = errors.As(err, &e)
	return
}
