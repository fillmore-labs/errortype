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

package qualname_test

import (
	"go/token"
	"go/types"
	"testing"

	. "fillmore-labs.com/errortype/internal/qualname"
)

func TestFuncOf(t *testing.T) {
	t.Parallel()

	pkg := types.NewPackage("example.com/testpkg", "testpkg")

	typeName := types.NewTypeName(token.NoPos, pkg, "MyType", nil)
	emptystruct := types.NewStruct(nil, nil)
	named := types.NewNamed(typeName, emptystruct, nil)
	aliasName := types.NewTypeName(token.NoPos, pkg, "MyAlias", nil)
	alias := types.NewAlias(aliasName, types.NewPointer(named))

	tests := [...]struct {
		name         string
		fun          *types.Func
		wantFuncName string
	}{
		{
			name: "simple function call",
			fun: func() *types.Func {
				sig := types.NewSignatureType(nil, nil, nil, nil, nil, false)

				return types.NewFunc(token.NoPos, pkg, "myFunc", sig)
			}(),
			wantFuncName: "example.com/testpkg.myFunc",
		},
		{
			name: "simple value method call",
			fun: func() *types.Func {
				recv := types.NewParam(token.NoPos, pkg, "", named)
				sig := types.NewSignatureType(recv, nil, nil, nil, nil, false)

				return types.NewFunc(token.NoPos, pkg, "myFunc", sig)
			}(),
			wantFuncName: "(example.com/testpkg.MyType).myFunc",
		},
		{
			name: "simple pointer method call",
			fun: func() *types.Func {
				recv := types.NewParam(token.NoPos, pkg, "", types.NewPointer(named))
				sig := types.NewSignatureType(recv, nil, nil, nil, nil, false)

				return types.NewFunc(token.NoPos, pkg, "myFunc", sig)
			}(),
			wantFuncName: "(example.com/testpkg.MyType).myFunc",
		},
		{
			name: "alias pointer method call",
			fun: func() *types.Func {
				recv := types.NewParam(token.NoPos, pkg, "", alias)
				sig := types.NewSignatureType(recv, nil, nil, nil, nil, false)

				return types.NewFunc(token.NoPos, pkg, "myFunc", sig)
			}(),
			wantFuncName: "(example.com/testpkg.MyType).myFunc",
		},
		{
			name: "interface method call",
			fun: func() *types.Func {
				sig := types.NewSignatureType(nil, nil, nil, nil, nil, false)
				iface := types.NewInterfaceType([]*types.Func{
					types.NewFunc(token.NoPos, pkg, "myFunc", sig),
				}, nil).Complete()

				return iface.Method(0)
			}(),
			wantFuncName: "(interface).myFunc",
		},
		{
			name: "function without package",
			fun: func() *types.Func {
				sig := types.NewSignatureType(nil, nil, nil, nil, nil, false)

				return types.NewFunc(token.NoPos, nil, "myFunc", sig)
			}(),
			wantFuncName: "myFunc",
		},
		{
			name: "method on type without package",
			fun: func() *types.Func {
				return types.Universe.Lookup("error").Type().Underlying().(*types.Interface).Method(0)
			}(),
			wantFuncName: "(error).Error",
		},
		{
			name: "invalid method call",
			fun: func() *types.Func {
				recv := types.NewParam(token.NoPos, pkg, "", emptystruct)
				sig := types.NewSignatureType(recv, nil, nil, nil, nil, false)

				return types.NewFunc(token.NoPos, pkg, "myFunc", sig)
			}(),
			wantFuncName: "(<invalid>).myFunc",
		},
		{
			name: "invalid pointer method call",
			fun: func() *types.Func {
				recv := types.NewParam(token.NoPos, pkg, "", types.NewPointer(emptystruct))
				sig := types.NewSignatureType(recv, nil, nil, nil, nil, false)

				return types.NewFunc(token.NoPos, pkg, "myFunc", sig)
			}(),
			wantFuncName: "(<invalid>).myFunc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if name := FuncOf(tt.fun); name.String() != tt.wantFuncName {
				t.Errorf("FuncOf() = %q, want %q", name, tt.wantFuncName)
			}
		})
	}
}

func TestFunc_UnmarshalText(t *testing.T) {
	t.Parallel()

	tests := [...]struct {
		name    string
		text    string
		want    Func
		wantErr bool
	}{
		// Methods with package path.
		{
			name: "method with package",
			text: "(encoding/json.Decoder).Decode",
			want: Func{Path: "encoding/json", LocalFunc: LocalFunc{Receiver: "Decoder", Name: "Decode"}},
		},
		// Pointer receiver (FullName compatibility).
		{
			name: "pointer receiver",
			text: "(*encoding/json.Decoder).Decode",
			want: Func{Path: "encoding/json", LocalFunc: LocalFunc{Receiver: "Decoder", Name: "Decode"}},
		},
		// Method without package path.
		{
			name: "method without package",
			text: "(MyType).Method",
			want: Func{LocalFunc: LocalFunc{Receiver: "MyType", Name: "Method"}},
		},
		// Pointer method without package path.
		{
			name: "pointer method without package",
			text: "(*MyType).Method",
			want: Func{LocalFunc: LocalFunc{Receiver: "MyType", Name: "Method"}},
		},
		// Interface method.
		{
			name: "interface method",
			text: "(interface).X",
			want: Func{LocalFunc: LocalFunc{Receiver: "interface", Name: "X"}},
		},
		// Regular function with package path.
		{
			name: "function with package",
			text: "fmt.Errorf",
			want: Func{Path: "fmt", LocalFunc: LocalFunc{Name: "Errorf"}},
		},
		// Function with dotted package path.
		{
			name: "function with dotted package",
			text: "example.com/pkg.New",
			want: Func{Path: "example.com/pkg", LocalFunc: LocalFunc{Name: "New"}},
		},
		// Bare function without package.
		{
			name: "bare function",
			text: "myFunc",
			want: Func{LocalFunc: LocalFunc{Name: "myFunc"}},
		},
		// Error cases.
		{name: "empty input", text: "", wantErr: true},
		{name: "unclosed paren", text: "(foo.Bar", wantErr: true},
		{name: "no dot after paren", text: "(foo)Method", wantErr: true},
		{name: "nothing after close paren dot", text: "(foo).", wantErr: true},
		{name: "empty receiver", text: "().Method", wantErr: true},
		{name: "star only receiver", text: "(*).Method", wantErr: true},
		{name: "empty path receiver", text: "(path.).Method", wantErr: true},
		{name: "trailing dot function", text: "pkg.", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var fn Func
			if err := fn.UnmarshalText([]byte(tt.text)); (err != nil) != tt.wantErr {
				t.Fatalf("UnmarshalText(%q) error = %v, wantErr %v", tt.text, err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if fn != tt.want {
				t.Errorf("UnmarshalText(%q) = %+v, want %+v", tt.text, fn, tt.want)
			}
		})
	}
}

func TestFunc_RoundTrip(t *testing.T) {
	t.Parallel()

	tests := [...]Func{
		{Path: "encoding/json", LocalFunc: LocalFunc{Receiver: "Decoder", Name: "Decode"}},
		{Path: "fmt", LocalFunc: LocalFunc{Name: "Errorf"}},
		{LocalFunc: LocalFunc{Name: "myFunc"}},
		{LocalFunc: LocalFunc{Receiver: "MyType", Name: "Method"}},
		{Path: "example.com/pkg", LocalFunc: LocalFunc{Receiver: "T", Name: "Do"}},
	}

	for _, original := range tests {
		t.Run(original.String(), func(t *testing.T) {
			t.Parallel()

			text, err := original.MarshalText()
			if err != nil {
				t.Fatalf("MarshalText() error = %v", err)
			}

			var decoded Func
			if err := decoded.UnmarshalText(text); err != nil {
				t.Fatalf("UnmarshalText(%q) error = %v", text, err)
			}

			if decoded != original {
				t.Errorf("round-trip failed: got %+v, want %+v", decoded, original)
			}
		})
	}
}
