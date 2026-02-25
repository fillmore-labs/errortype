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

package analyze

import (
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/edge"
	"golang.org/x/tools/go/ast/inspector"

	"fillmore-labs.com/errortype/internal/analyze/usage"
	"fillmore-labs.com/errortype/internal/astutil"
	"fillmore-labs.com/errortype/internal/errorlike"
	"fillmore-labs.com/errortype/internal/settings"
	"fillmore-labs.com/errortype/internal/typeutil"
)

func (p Pass) handleMethodDecl(c inspector.Cursor, f *ast.FuncDecl, file *ast.File) {
	fun, ok := p.TypesInfo.Defs[f.Name].(*types.Func)
	if !ok { // should not happen
		return
	}

	sig := fun.Signature()
	if sig.Recv() == nil { // should not happen, checked before
		return
	}

	tn, ptr, ok := typeutil.CanonicalTypeNameOf(sig.Recv().Type())
	if !ok { // should not happen
		return
	}

	prop, ok := p.ErrorUsage[tn]
	if !ok {
		return // not an error type
	}

	if astutil.HasNoLint(f.Doc, p.Analyzer.Name) {
		return
	}

	method, isErrorMethod := errorlike.MethodNamed(f.Name.Name)
	if isErrorMethod && !method.MatchSig(sig) {
		// also found by golang.org/x/tools/go/analysis/passes/stdmethods.
		p.ReportRangef(f.Type, "Method %q has the wrong signature (et:sig)", fun.Name())

		return
	}

	p.checkReceiver(f, file, tn, ptr, prop&usage.ExpectedMask, method, isErrorMethod)

	// check for an "Is" method declaration and hand over to body analysis.
	if isErrorMethod && method.IsIs() && f.Body != nil {
		param := singleParam(p.TypesInfo, f.Type.Params)
		if param == nil {
			return // An "Is" method without named target parameter is legal, but rare.
		}

		body := c.ChildAt(edge.FuncDecl_Body, -1)
		p.handleIs(body, param)

		if ptr && typeutil.ZeroSized(tn.Type()) {
			return // make an exception for pointers to zero-sized types.
		}

		p.checkAssertionQuery(f, tn)
	}
}

func (p Pass) checkAssertionQuery(f *ast.FuncDecl, tn *types.TypeName) {
	q, ok := p.matchAssertionQuery(f.Type, f.Body)
	if !ok {
		return
	}

	if atn, _, ok := typeutil.CanonicalTypeNameOf(q.assertedType); !ok || atn != tn {
		return
	}

	// If aliased, use the alias name
	atn, ptr, ok := typeutil.TypeNameOf(q.assertedType)
	if !ok {
		return
	}

	name := atn.Name()
	if ptr {
		name = "*" + name
	}

	p.Report(analysis.Diagnostic{
		Pos:     f.Type.Pos(),
		End:     f.Type.End(),
		Message: fmt.Sprintf("Is method implementation makes errors.Is act as a type check; remove it and use errors.AsType[%s] at call sites (et:ias)", name),
	})
}

// checkReceiver reports a method receiver that does not match the usage of the error type tn.
//
// Always reported are pointer receivers of Unwrap, Is and As on value errors, which the value
// lacks, and a value receiver of Error on a pointer error with a pointer-only Unwrap, Is or As:
// then the value implements error without those methods, and go vet (in modules at go 1.27 or
// later) asks to wrap the value with %w instead of the pointer.
//
// With [settings.OptionReceivers], the check follows the method sets:
//   - Value errors are reported for pointer receivers on any exported method, like Timeout of
//     net.Error, which the value lacks when queried through an interface. Unexported methods
//     can only satisfy interfaces of the same package, and decoder methods need a pointer
//     receiver to write to the value, see [decoderMethods].
//   - Pointer errors are reported for a value receiver on Error, which makes the value an error,
//     too. Other value-receiver methods are in the method set of the pointer, so nothing is hidden.
//
// In files at Go 1.27 or later, a value receiver on Error of a pointer error is reported without
// [settings.OptionReceivers]: there, go vet reports wrapping the pointer with %w anyway.
//
// An Error method promoted from an embedded field is not reported: nobody chose its receiver, so it is
// sufficient that the type is used consistently.
func (p Pass) checkReceiver(f *ast.FuncDecl, file *ast.File, tn *types.TypeName, ptr bool, expected usage.Usage, method errorlike.Method, isErrorMethod bool) {
	switch name := f.Name.Name; {
	case ptr && expected == usage.ValueExpected:
		if isErrorMethod && !method.IsError() ||
			p.Has(settings.OptionReceivers) && ast.IsExported(name) && !decoderMethod(name) {
			p.ReportRangef(f.Recv, "Method %q should be implemented with a value receiver (\"func (%s) %s\"), not a pointer receiver (et:rcv)", name, tn.Name(), name)
		}

	case !ptr && expected == usage.PointerExpected && isErrorMethod && method.IsError():
		if pointerMethod := pointerOnlyMethod(tn.Type()); pointerMethod != "" {
			p.ReportRangef(f.Recv, "Method %q should be implemented with a pointer receiver (\"func (*%s) %s\"), like %q (et:rcv+)", name, tn.Name(), name, pointerMethod)

			return
		}

		if p.Has(settings.OptionReceivers) || goVersionAtLeast(p.TypesInfo, file, "go1.27") {
			p.ReportRangef(f.Recv, "Method %q should be implemented with a pointer receiver (\"func (*%s) %s\"), not a value receiver (et:rcv+)", name, tn.Name(), name)
		}
	}
}

func decoderMethod(m string) bool {
	_, ok := decoderMethods[m]

	return ok
}

// decoderMethods are methods that decode into their receiver, so they need a pointer receiver,
// even on value errors. They are called on an addressable variable, not through the error interface.
//
// This list is preliminary. It is matched by name only, since these names are unlikely to be used
// for anything else on error types.
var decoderMethods = map[string]struct{}{
	"UnmarshalJSON":     {}, // encoding/json.Unmarshaler
	"UnmarshalJSONFrom": {}, // encoding/json/v2.UnmarshalerFrom
	"UnmarshalText":     {}, // encoding.TextUnmarshaler
	"UnmarshalBinary":   {}, // encoding.BinaryUnmarshaler
	"UnmarshalYAML":     {}, // YAML libraries
	"UnmarshalTOML":     {}, // TOML libraries
	"UnmarshalXML":      {}, // encoding/xml.Unmarshaler
	"GobDecode":         {}, // encoding/gob.GobDecoder
	"DecodeMsgpack":     {}, // msgpack libraries
	"Scan":              {}, // database/sql.Scanner
	"Set":               {}, // flag.Value, pflag.Value
}

// pointerOnlyMethod returns the name of an Unwrap, Is or As method that is in the method set of *T,
// but not of T, or "" if there is none.
func pointerOnlyMethod(typ types.Type) string {
	for _, m := range [...]errorlike.Method{errorlike.UnwrapMethod, errorlike.IsMethod, errorlike.AsMethod} {
		if m.HasMethod(typ, true) && !m.HasMethod(typ, false) {
			return m.Name()
		}
	}

	return ""
}
