# Errortype

[![Go Reference](https://pkg.go.dev/badge/fillmore-labs.com/errortype.svg)](https://pkg.go.dev/fillmore-labs.com/errortype)
[![Test](https://github.com/fillmore-labs/errortype/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/fillmore-labs/errortype/actions/workflows/test.yml)
[![CodeQL](https://github.com/fillmore-labs/errortype/actions/workflows/github-code-scanning/codeql/badge.svg?branch=main)](https://github.com/fillmore-labs/errortype/actions/workflows/github-code-scanning/codeql)
[![Coverage](https://codecov.io/gh/fillmore-labs/errortype/branch/main/graph/badge.svg?token=MMLHL14ZP6)](https://codecov.io/gh/fillmore-labs/errortype)
[![Go Report Card](https://goreportcard.com/badge/fillmore-labs.com/errortype)](https://goreportcard.com/report/fillmore-labs.com/errortype)
[![License](https://img.shields.io/github/license/fillmore-labs/errortype)](https://www.apache.org/licenses/LICENSE-2.0)

`errortype` is a Go static analysis tool (linter) that detects inconsistent usage of custom error types as pointers
versus values. It helps prevent subtle bugs by ensuring error types are used consistently throughout your codebase,
eliminating a common source of bugs in error handling logic.

## Motivation

In Go, error types can be designed for use as either values or pointers. Inconsistent use can lead to subtle,
hard-to-find bugs.

Consider the following code, which attempts to provide a more specific error message for an incorrect AES key size
([Go Playground](https://go.dev/play/p/m4SEPqkZ2Zu)):

```go
package main

import (
	"crypto/aes"
	"errors"
	"fmt"
)

func main() {
	key := []byte("My kung fu is better than yours")
	_, err := aes.NewCipher(key)

	var kse *aes.KeySizeError
	if errors.As(err, &kse) {
		fmt.Printf("AES keys must be 16, 24 or 32 bytes long, got %d bytes.\n", kse)
	} else if err != nil {
		fmt.Println(err)
	}
}
```

This code doesn't work as intended because `aes.KeySizeError` is designed to be used as a value, not a pointer. As
written, the code prints the generic error message instead of the custom one.

Changing line 13 to “`var kse aes.KeySizeError`” fixes the issue, and the program correctly prints
“`AES keys must be 16, 24 or 32 bytes long, got 31 bytes.`”

### Why Consistency Matters

Inconsistent usage of error types can lead to hard-to-spot bugs. `errortype` prevents these issues by automatically
detecting the intended usage of error types and reporting inconsistencies in:

- Function return values
- Type assertions and type switches
- Calls to `errors.As` and similar functions (e.g., from [`testify`](https://pkg.go.dev/github.com/stretchr/testify))

In the above example, `errortype .` would report:

```console
/path/to/your/source/main.go:14:20: Target for value error "crypto/aes.KeySizeError" is a pointer-to-pointer, use a pointer to a value instead: "var kse aes.KeySizeError; ... errors.As(err, &kse)". (et:err)
```

This message suggests changing the variable declaration to “`var kse aes.KeySizeError`”, which corrects the bug.

### Style Checks

Go's syntax allows for some “clever” but potentially confusing constructs with `errors.As`. Consider the following:

```go
  if errors.As(err, &MyError{}) { /* ... */ } // Is this checking for a pointer or a value error?
```

While this is valid code, it can be misleading. The expression `&MyError{}` creates a pointer to a struct literal. When
passed to `errors.As`, it checks if `err` wraps a `MyError` _value_, not a pointer. This is easily misread, obfuscating
the check's true intent. To improve clarity and prevent misinterpretation, `errortype` encourages a more explicit and
readable style:

```go
  var e MyError // Or "var e *MyError" for pointer errors
  if errors.As(err, &e) { /* ... */ } // The clear, recommended style
```

This longer form is unambiguous and clearly states the type being checked for. `errortype` emits an `(et:sty)` warning
for constructs where the target argument of an `errors.As`-like function is not an address of a variable.

### Linter Scope

The primary goal of this linter is to enforce consistent usage of error types.

Good error type design is out of scope. While `errortype` promotes a consistent style, a broader refactor of your error
handling strategy may sometimes be the better solution.

## Getting Started

### Installation

Choose one of the following installation methods:

#### Homebrew

```console
brew install fillmore-labs/tap/errortype
```

#### Go

```console
go install fillmore-labs.com/errortype@latest
```

#### Eget

[Install `eget`](https://github.com/zyedidia/eget?tab=readme-ov-file#how-to-get-eget), then

```console
eget fillmore-labs/errortype
```

## Usage

To analyze your entire project, run:

```console
errortype ./...
```

### Command-Line Flags

Usage: `errortype [-flag] [package]`

`errortype` supports the following flags:

- **-overrides** `<filename>`: Read type overrides from the specified YAML file. See the
  [“Overrides File”](#overrides-file) section for more details.
- **-suggest** `<filename>`: Append suggestions for an override file. Use `-` for standard output.
- **-stylecheck**: Check whether targets of errors.As-like functions are address operators on variables (default: true).
- **-c** `<N>`: Display N lines of context around each issue (default: -1 for no context, 0 for only the offending
  line).
- **-test**: Analyze test files in addition to source files (default: true).
- **-heuristics**: (Experimental) List of heuristics used (default: "usage,receivers", "off" to disable).
- **-debug**: (Experimental) Output information for debugging.

## How Intended Usage is Detected

The linter determines an error type's intended use (pointer vs. value) by analyzing the package where the error is
**defined**. It uses the following order of precedence:

1. **Package-Level Variable Assignments**: If present, `var _ error = ...` assignments are used as explicit declarations
   of intent.

   ```go
   var _ error = ValueError{}         // Determines ValueError is a "value" type.

   var _ error = (*PointerError)(nil) // Determines PointerError is a "pointer" type.
   ```

2. **Overrides**: User-defined overrides (see [below](#overrides-file)) are applied next, overriding any previously
   detected usage.

3. **Usage within Functions**: If still undecided, the linter analyzes usage within top-level functions (e.g., in
   `return` statements or type assertions). Consistent usage can determine the type.

   ```go
   return ValueError{} // Suggests value type

   if _, ok := err.(*PointerError); ok { /* ... */ } // Suggests pointer type
   ```

   Note: This heuristic is a fallback and should not be relied upon for defining a type's contract.

4. **Consistent Method Receivers**: As a final heuristic, if all methods on a type have a consistent receiver (all-value
   or all-pointer), that style is used.

### Limitations

If `errortype` cannot determine the intended usage (e.g., for types that embed the `error` interface without consistent
receivers), it reports an `et:emb` diagnostic. This can be resolved [using an override](#overriding-detected-types).

### Designing Linter-Friendly Packages

To make an error type's intended usage explicit and ensure `errortype` can automatically determine it, add a
package-level variable assignment in the package where the error is defined:

```go
// In your package, explicitly declare the intended usage.
var _ error = ValueError{}

var _ error = (*PointerError)(nil)
```

### Overriding Detected Types

If the linter reports types from an imported package with ambiguous or inconsistent usage, you can guide the linter in
two ways:

1. **Local Override**: For a one-off fix within a single package, add a `var` block to a source file in that package.
   This overrides the detected usage for that type _within this package only_.

   ```go
   // In your code, force a specific usage for an imported type.
   var _ error = imported.ValueError{}

   var _ error = (*imported.PointerError)(nil)
   ```

2. **Global Override File**: For project-wide overrides, use an `errortypes.yaml` file.

## Overrides File

You can generate a sample override file with the `-suggest` flag. This file will contain a list of types that require a
decision:

```console
errortype -suggest=errortypes.yaml ./...
```

This command creates (or appends to) `errortypes.yaml` with the following structure:

```yaml
# Override types for your.path/package
---
pointer: # Types that should always be used as pointers
  - imported.path/one.PointerOverride

value: # Types that should always be used as values
  - imported.path/two.ValueOverride

suppress: # Types to completely ignore during analysis
  - imported.path/one.ErrorToIgnore

inconsistent: # Types that are used inconsistently (generated by -suggest)
  - imported.path/two.InconsistentUsage
```

The `inconsistent` section is only generated by `-suggest` and is ignored by the linter. You can review these entries
and move them to the `pointer`, `value`, or `suppress` sections as appropriate.

Once your `errortypes.yaml` file is configured, use it with the `-overrides` flag:

```console
errortype -overrides=errortypes.yaml ./...
```

This instructs the linter to use your specified configuration, resolving ambiguities and suppressing noise from types
you wish to ignore.

**Note:** Always review suggestions before adding them to your overrides file. A suggestion makes your code consistent
with how the type is _used in your package_, but this may conflict with how the type was _designed_ to be used in its
defining package. When possible, fixing the inconsistency by refactoring the code is preferable to forcing an override.

### Overrides vs. Autodetection

It's important to understand the difference between autodetection and overrides.

- **Autodetection** runs on the package where an error type is **defined**, see
  “[How Intended Usage is Detected](#how-intended-usage-is-detected)”. This is the ideal place to establish the intended
  usage.

- **Overrides** are based on the usage within **your** code. They force a specific pointer or value style, overriding
  what was detected in the defining package.

**Every suggestion should be reviewed before being used as an override.** An `inconsistent` usage report may indicate a
genuine opportunity to refactor and improve your error handling. When possible, it is always better to improve detection
in the defining package by making the usage explicit, see
“[Designing Linter-Friendly Packages](#designing-linter-friendly-packages).”

## Diagnostic Code Reference

`errortype` uses short codes to categorize the issues it finds.

- **`et:ret` (Return Mismatch)**: An error type is returned incorrectly.

  ```go
  return &ValueError{} // Returning a value error as a pointer
  ```

- **`et:ast` (Assertion Mismatch)**: An error type is used incorrectly in a type assertion or type switch.

  ```go
  target, ok := err.(*ValueError) // Asserting a value error to a pointer type
  ```

- **`et:err` (Argument Mismatch)**: An error type is passed incorrectly as a target to an `errors.As`-like function.

  ```go
  var target *ValueError
  errors.As(err, &target) // The target for a value error is a pointer-to-pointer
  ```

- **`et:emb` (Embedded/Ambiguous Usage)**: The linter could not determine if an error is a pointer or value type. This
  is common for types that embed the `error` interface or have mixed usage in the defining package.

  ```go
  type AmbiguousError struct{ error }
  // ...
  return AmbiguousError{err}
  ```

  See “[Overriding Detected Types](#overriding-detected-types).”

- **`et:sty` (Style Mismatch)**: The target argument to an `errors.As`-like function is not an address operation on a
  variable.

  ```go
  if ee := new(*exec.ExitError); !errors.As(err, ee) { /* ... */ }
  ```

- **`et:arg` (Invalid Argument)**: The target argument to an `errors.As`-like function is invalid (e.g., not a pointer
  to a type implementing error).

  ```go
  var target net.ParseError    // target is not an error type
  if errors.As(err, &target) { /* ... */ }
  ```

  This is also flagged by the standard [`errorsas`](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/errorsas)
  linter.

## Integration

This linter is in an early phase and is currently usable only from the command line. Other integrations are planned as
the logic stabilizes.

## Real-World Examples

See [this blog post](https://blog.fillmore-labs.com/posts/errors-1/) for why `errortype` is useful.

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.
