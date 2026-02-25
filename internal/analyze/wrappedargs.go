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

package analyze

import (
	"iter"
	"strings"
	"unicode/utf8"
)

// AllWrappedArgs parses a printf-style format string and returns the zero-based
// indices of the operands wrapped by %w verbs, in order of appearance.
//
// Parsing mirrors fmt's doPrintf: flags, width and precision are taken into
// account, a width or precision given as "*" consumes an operand, an explicit
// argument index like "[2]" repositions the operand counter and "%%" consumes
// no operand. Verbs without a usable operand are dropped, mirroring fmt, which
// reports "%!w(MISSING)" or "%!w(BADINDEX)" instead of wrapping.
func AllWrappedArgs(format string, numArgs int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for s := newFormatScanner(format, numArgs); ; {
			switch verb, argNum, ok := s.next(); {
			case !ok:
				return

			case verb != 'w':
				continue

			case !yield(argNum):
				return
			}
		}
	}
}

// A formatScanner scans a printf-style format string, replicating the parsing
// logic of fmt's doPrintf.
type formatScanner struct {
	format  string
	numArgs int  // number of operands available to the conversions
	pos     int  // current scan position in format
	argNum  int  // index of the operand consumed by the next conversion
	good    bool // whether the current conversion has a usable operand index
}

func newFormatScanner(format string, numArgs int) formatScanner {
	return formatScanner{format: format, numArgs: numArgs}
}

// next returns the verb of the next conversion that consumes an operand,
// together with the zero-based index of that operand. Conversions that consume
// no operand ("%%", bad argument indexes, missing operands) are skipped. The
// result ok is false once the format string is exhausted.
func (s *formatScanner) next() (verb rune, argNum int, ok bool) {
	for s.nextPercent() {
		if verb, argNum, consumed := s.conversion(); consumed {
			return verb, argNum, true
		}
	}

	return 0, 0, false
}

// nextPercent advances after the next '%' and returns true when successful.
func (s *formatScanner) nextPercent() bool {
	if s.pos >= len(s.format) {
		return false
	}

	i := strings.IndexByte(s.format[s.pos:], '%')
	if i < 0 {
		return false
	}

	s.pos += i + 1 // Skip the '%'.

	return true
}

// conversion parses a single conversion after a '%' and returns its verb and
// the index of the operand it consumes. The result consumed is false when the
// conversion consumes no operand.
func (s *formatScanner) conversion() (verb rune, argNum int, consumed bool) {
	s.good = true

	// Flags do not affect operand consumption.
	for s.pos < len(s.format) && isFlag(s.format[s.pos]) {
		s.pos++
	}

	// An explicit argument index may precede the width, the precision and the verb.
	afterIndex := s.argNumber()

	// A width given as "*" consumes an operand.
	if s.consume('*') {
		s.skipOperand()

		afterIndex = false
	} else if s.number() && afterIndex {
		s.good = false // e.g. "%[2]4d"
	}

	// A precision given as ".*" consumes an operand, too.
	if s.pos+1 < len(s.format) && s.format[s.pos] == '.' {
		s.pos++

		if afterIndex {
			s.good = false // e.g. "%[2].4d"
		}

		afterIndex = s.argNumber()
		if s.consume('*') {
			s.skipOperand()

			afterIndex = false
		} else {
			s.number()
		}
	}

	if !afterIndex {
		s.argNumber()
	}

	if s.pos >= len(s.format) {
		return 0, 0, false // Missing verb.
	}

	verb, size := utf8.DecodeRuneInString(s.format[s.pos:])
	s.pos += size

	// "%%" does not consume an operand; neither do conversions with a bad
	// argument index or with no operand left.
	if verb == '%' || !s.good || s.argNum >= s.numArgs {
		return verb, 0, false
	}

	argNum = s.argNum
	s.argNum++

	return verb, argNum, true
}

// argNumber parses an explicit argument index like "[2]", repositioning argNum
// to the one-based index given. It reports whether the scan position is
// directly behind a well-formed index; a malformed or out-of-range index
// invalidates the current conversion.
func (s *formatScanner) argNumber() (afterIndex bool) {
	if s.pos >= len(s.format) || s.format[s.pos] != '[' {
		return false
	}

	index, width, parsed := parseArgNumber(s.format[s.pos:])
	s.pos += width

	if parsed && 0 <= index && index < s.numArgs {
		s.argNum = index

		return true
	}

	s.good = false

	return parsed
}

// parseArgNumber returns the zero-based value of the bracketed argument index
// at the start of format and the number of bytes it occupies. It mirrors fmt's
// parseArgNumber.
func parseArgNumber(format string) (index, width int, ok bool) {
	// There must be at least 3 bytes: [n].
	if len(format) < 3 {
		return 0, 1, false
	}

	// Find the closing bracket.
	i := strings.IndexByte(format[1:], ']')
	if i < 0 {
		return 0, 1, false // No closing bracket.
	}

	value, isNum, numLen := parseNum(format[1 : i+1])
	if !isNum || numLen != i {
		return 0, i + 2, false
	}

	return value - 1, i + 2, true // Argument indexes are one-based.
}

// consume advances the scan position over the byte c if it is next.
func (s *formatScanner) consume(c byte) bool {
	if s.pos >= len(s.format) || s.format[s.pos] != c {
		return false
	}

	s.pos++

	return true
}

// number skips the decimal number at the current position, used as a width or
// precision, and reports whether one was present.
func (s *formatScanner) number() bool {
	_, isNum, numLen := parseNum(s.format[s.pos:])
	s.pos += numLen

	return isNum
}

// skipOperand consumes the operand of a width or precision given as "*".
func (s *formatScanner) skipOperand() {
	if s.argNum < s.numArgs {
		s.argNum++
	}
}

// maxNum limits the widths, precisions and argument indexes accepted in
// conversions, mirroring fmt's tooLarge.
const maxNum = 1_000_000

// parseNum parses the decimal number at format, mirroring fmt's
// parsenum. Oversized numbers abort the scan by jumping to end.
func parseNum(format string) (num int, isNum bool, numLen int) {
	end := len(format)
	for numLen = 0; numLen < end && '0' <= format[numLen] && format[numLen] <= '9'; numLen++ {
		if num > maxNum {
			return 0, false, end
		}

		num = num*10 + int(format[numLen]-'0')
		isNum = true
	}

	return num, isNum, numLen
}

// isFlag reports whether c is a printf flag character.
func isFlag(c byte) bool {
	switch c {
	case '#', '0', '+', '-', ' ':
		return true

	default:
		return false
	}
}
