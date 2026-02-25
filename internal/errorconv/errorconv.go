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

package errorconv

import (
	"go/ast"
	"go/types"
	"iter"

	"fillmore-labs.com/errortype/internal/errorlike"
	"fillmore-labs.com/errortype/internal/typeutil"
)

// CallArgs yields (argType, argExpr) pairs for arguments in call that are
// converted to an error-like interface parameter, including explicit interface
// conversions such as error(e) and builtins with recorded signatures such as
// append(errs, e).
//
// For a multivalued call f(g()), argExpr is g() for every tuple element passed
// to an error-like parameter. A slice spread with ... is passed as is and is not
// yielded. The callee may be a value of type parameter type with a core function
// type.
//
// argType is not filtered: it may already be an interface (an identity
// conversion) or a type parameter, since a type parameter constrained by an
// error-like interface counts as error-like itself.
func CallArgs(info *types.Info, call *ast.CallExpr) iter.Seq2[types.Type, ast.Expr] {
	return func(yield func(types.Type, ast.Expr) bool) {
		if len(call.Args) == 0 {
			return
		}

		funTV, ok := info.Types[call.Fun]
		if !ok {
			return
		}

		if funTV.IsType() {
			// Explicit conversion to an error interface, e.g., error(e).
			if errorlike.IsInterface(funTV.Type) && len(call.Args) == 1 {
				if argTV, ok := info.Types[call.Args[0]]; ok {
					yield(argTV.Type, call.Args[0])
				}
			}

			return
		}

		sig, ok := typeutil.CoreType(funTV.Type).(*types.Signature)
		if !ok || sig.Params().Len() == 0 {
			return
		}

		// Multivalued call: fn(g()) where g() returns a tuple (including variadic calls like errors.Join(pair())).
		if len(call.Args) == 1 && !call.Ellipsis.IsValid() {
			if argTV, ok := info.Types[call.Args[0]]; ok {
				if tuple, ok := argTV.Type.(*types.Tuple); ok {
					yieldTupleCallArgs(sig, call.Args[0], tuple, yield)
					return
				}
			}
		}

		yieldRegularCallArgs(info, sig, call, yield)
	}
}

// yieldTupleCallArgs yields the tuple element types of the single argument arg
// that are passed to error-like parameters of sig, each paired with arg.
func yieldTupleCallArgs(
	sig *types.Signature,
	arg ast.Expr,
	tuple *types.Tuple,
	yield func(types.Type, ast.Expr) bool,
) {
	for i := range tuple.Len() {
		paramType := paramTypeAt(sig, i)
		if paramType == nil {
			return
		}

		if !errorlike.IsInterface(paramType) {
			continue
		}

		if !yield(tuple.At(i).Type(), arg) {
			return
		}
	}
}

// yieldRegularCallArgs yields the arguments of call passed to error-like
// parameters of sig.
func yieldRegularCallArgs(
	info *types.Info,
	sig *types.Signature,
	call *ast.CallExpr,
	yield func(types.Type, ast.Expr) bool,
) {
	args := call.Args
	if call.Ellipsis.IsValid() {
		// f(a, s...): the slice s is passed as the variadic parameter as is.
		args = args[:len(args)-1]
	}

	for i, arg := range args {
		paramType := paramTypeAt(sig, i)
		if paramType == nil {
			return
		}

		if !errorlike.IsInterface(paramType) {
			continue
		}

		if argTV, ok := info.Types[arg]; ok {
			if !yield(argTV.Type, arg) {
				return
			}
		}
	}
}

// paramTypeAt returns the type of the parameter that receives argument i in a
// call of sig. Arguments matching a variadic parameter ...T receive type T.
// It returns nil when there are more arguments than parameters.
func paramTypeAt(sig *types.Signature, i int) types.Type {
	params := sig.Params()

	switch last := params.Len() - 1; {
	case sig.Variadic() && i >= last:
		// The variadic parameter is a slice []T, except for the builtin
		// append([]byte, string...), whose spread argument callers don't ask for.
		slice, ok := params.At(last).Type().(*types.Slice)
		if !ok {
			return nil
		}

		return slice.Elem()

	case i <= last:
		return params.At(i).Type()

	default:
		return nil
	}
}

// CompositeLitElts yields (valType, valExpr) pairs for elements, map keys/values,
// or struct fields in lit whose target type is an error-like interface.
//
// It handles literals with elided type &T, as in []*T{{...}}, and literals of
// type parameter type with a core slice, array, map or struct type.
func CompositeLitElts(info *types.Info, lit *ast.CompositeLit) iter.Seq2[types.Type, ast.Expr] {
	return func(yield func(types.Type, ast.Expr) bool) {
		if len(lit.Elts) == 0 {
			return
		}

		litTV, ok := info.Types[lit]
		if !ok {
			return
		}

		// The recorded type of an elided &T{...} is *T.
		litType := litTV.Type
		if ptr, ok := types.Unalias(litType).(*types.Pointer); ok {
			litType = ptr.Elem()
		}

		switch t := typeutil.CoreType(litType).(type) {
		case *types.Slice:
			yieldSeqElts(info, lit.Elts, t.Elem(), yield)

		case *types.Array:
			yieldSeqElts(info, lit.Elts, t.Elem(), yield)

		case *types.Map:
			yieldMapElts(info, lit.Elts, t, yield)

		case *types.Struct:
			yieldStructElts(info, lit.Elts, t, yield)
		}
	}
}

// yieldSeqElts yields the slice or array elements elts when elemType is error-like.
func yieldSeqElts(
	info *types.Info,
	elts []ast.Expr,
	elemType types.Type,
	yield func(types.Type, ast.Expr) bool,
) {
	if !errorlike.IsInterface(elemType) {
		return
	}

	for _, elt := range elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			elt = kv.Value
		}

		if eltTV, ok := info.Types[elt]; ok {
			if !yield(eltTV.Type, elt) {
				return
			}
		}
	}
}

// yieldMapElts yields the keys and values of the map elements elts whose
// corresponding map key or element type is error-like.
func yieldMapElts(
	info *types.Info,
	elts []ast.Expr,
	m *types.Map,
	yield func(types.Type, ast.Expr) bool,
) {
	keyIsErr := errorlike.IsInterface(m.Key())
	elemIsErr := errorlike.IsInterface(m.Elem())

	if !keyIsErr && !elemIsErr {
		return
	}

	for _, elt := range elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		if keyIsErr {
			if keyTV, ok := info.Types[kv.Key]; ok {
				if !yield(keyTV.Type, kv.Key) {
					return
				}
			}
		}

		if elemIsErr {
			if valTV, ok := info.Types[kv.Value]; ok {
				if !yield(valTV.Type, kv.Value) {
					return
				}
			}
		}
	}
}

// yieldStructElts yields the keyed or positional field values elts of struct s
// whose field type is error-like.
func yieldStructElts(
	info *types.Info,
	elts []ast.Expr,
	s *types.Struct,
	yield func(types.Type, ast.Expr) bool,
) {
	for i, elt := range elts {
		var field *types.Var

		valExpr := elt
		if kv, isKeyed := elt.(*ast.KeyValueExpr); isKeyed {
			if id, isIdent := kv.Key.(*ast.Ident); isIdent {
				field, _ = info.Uses[id].(*types.Var)
			}

			valExpr = kv.Value
		} else if i < s.NumFields() {
			field = s.Field(i)
		}

		if field == nil || !errorlike.IsInterface(field.Type()) {
			continue
		}

		if valTV, ok := info.Types[valExpr]; ok {
			if !yield(valTV.Type, valExpr) {
				return
			}
		}
	}
}

// SendValue returns (valType, send.Value, true) when send targets a channel
// with an error-like element type and send.Value has recorded type information.
// The channel may be of type parameter type with a core channel type.
func SendValue(info *types.Info, send *ast.SendStmt) (types.Type, ast.Expr, bool) {
	chanTV, ok := info.Types[send.Chan]
	if !ok {
		return nil, nil, false
	}

	if ch, ok := typeutil.CoreType(chanTV.Type).(*types.Chan); !ok || !errorlike.IsInterface(ch.Elem()) {
		return nil, nil, false
	}

	valTV, ok := info.Types[send.Value]
	if !ok {
		return nil, nil, false
	}

	return valTV.Type, send.Value, true
}
