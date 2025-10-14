# Errortype

[![Go Reference](https://pkg.go.dev/badge/fillmore-labs.com/errortype.svg)](https://pkg.go.dev/fillmore-labs.com/errortype)
[![Test](https://github.com/fillmore-labs/errortype/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/fillmore-labs/errortype/actions/workflows/test.yml)
[![CodeQL](https://github.com/fillmore-labs/errortype/actions/workflows/github-code-scanning/codeql/badge.svg?branch=main)](https://github.com/fillmore-labs/errortype/actions/workflows/github-code-scanning/codeql)
[![Coverage](https://codecov.io/gh/fillmore-labs/errortype/branch/main/graph/badge.svg?token=MMLHL14ZP6)](https://codecov.io/gh/fillmore-labs/errortype)
[![Go Report Card](https://goreportcard.com/badge/fillmore-labs.com/errortype)](https://goreportcard.com/report/fillmore-labs.com/errortype)
[![Codeberg CI](https://ci.codeberg.org/api/badges/15305/status.svg?branch=main)](https://ci.codeberg.org/repos/15305/branches/main)
[![License](https://img.shields.io/github/license/fillmore-labs/errortype)](https://www.apache.org/licenses/LICENSE-2.0)

`errortype` is a Go static analysis tool (linter) that helps prevent subtle bugs in error handling and other areas. It
performs two main checks:

1. **Inconsistent Error Type Usage**: It analyzes function return values, type assertions, and calls to functions like
   `errors.As` to ensure that custom error types are used consistently as either pointers or values.

2. **Pointless Comparisons**: It detects comparisons against the address of a newly created value (like
   `errors.Is(err, &url.Error{})`), which are almost always incorrect.

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

[Install `eget`](https://github.com/zyedidia/eget#how-to-get-eget), then

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

- **-overrides** `<filename>`: Read type overrides from the specified YAML file, see the
  _[“Override File”](#override-file)_ section for more details.
- **-suggest** `<filename>`: Append suggestions to an override file for review. Use `-` for standard output.
- **-check-is**: Suppress diagnostics on `errors.Is` if the compared type has an `Is(error) bool` method (default:
  true).
- **-deep-is-check**: (Experimental) When checking an `Is(error) bool` method, diagnose any call to an unwrapping
  function (like `errors.Is` or `errors.As`). By default, only calls that use the `target` parameter are flagged
  (default: false).
- **-style-check**: Check for confusing uses of `errors.As` (default: true).
- **-unchecked-assert**: Diagnose unchecked type asserts on errors (default: false).
- **-c** `<N>`: Display N lines of context around each issue (default: -1 for no context, 0 for only the offending
  line).
- **-test**: Analyze test files in addition to source files (default: true).
- **-heuristics**: `<list>` (Experimental) List of heuristics used (default: “var,usage,receivers”, “off” to disable).
- **-tracetypes**: `<regex>` (Experimental) Trace error type detection in packages with names matching the regular
  expression.

## Inconsistent Error Type Usage

One of the most common and subtle error-handling bugs in Go occurs when error types are used inconsistently — sometimes
as values and sometimes as pointers. This inconsistency can lead to subtle, hard-to-find bugs.

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
		fmt.Printf("AES keys must be 16, 24, or 32 bytes long, got %d bytes.\n", kse)
	} else if err != nil {
		fmt.Println(err)
	}
}
```

This code doesn't work as intended because `aes.KeySizeError` is designed to be used as a value, not a pointer. As
written, the code prints the generic error message instead of the custom one.

Changing line 13 to `var kse aes.KeySizeError` fixes the issue, and the program correctly prints _“AES keys must be 16,
24, or 32 bytes long, got 31 bytes.”_

### How `errortype` Helps

The `errortype` linter prevents these issues by automatically detecting the intended usage of error types and reporting
inconsistencies in:

- Function return values
- Type assertions and type switches
- Calls to `errors.As` and similar functions (e.g., from [`testify`](https://pkg.go.dev/github.com/stretchr/testify))

In the above example, `errortype .` would report:

```console
.../main.go:14:20: Target for value error "crypto/aes.KeySizeError" ⏎
    is a pointer-to-pointer, use a pointer to a value instead: ⏎
    "var kse aes.KeySizeError; ... errors.As(err, &kse)". (et:err)
```

The message suggests changing the variable declaration to `var kse aes.KeySizeError`, which corrects the bug.

### How Intended Usage Is Detected

The linter determines an error type's intended use _(pointer vs. value)_ by analyzing the package where the error is
**defined**. It uses the following order of precedence:

1. **Overrides** (the highest priority): User-defined overrides (see [below](#override-file)) are applied, overriding
   any detected usage.

2. **`Unwrap` related methods**: If present, `Is`, `As` and `Unwrap` methods with pointer receiver would be not visible
   from a value usage. If an error type would be used as a value, methods with pointer receivers are not in its
   [method set](https://go.dev/ref/spec#Method_sets).

   ```go
   func (e *PointerError) Unwrap error { /* ...*/ } // Unwrap is only visible from error(&PointerError{}).
   ```

3. **Package-Level Variable Assignments**: If present, `var _ error = ...` assignments are used as explicit declarations
   of intent.

   ```go
   var _ error = ValueError{}         // Determines ValueError is a "value" type.

   var _ error = (*PointerError)(nil) // Determines PointerError is a "pointer" type.
   ```

4. **Usage Within Functions**: If still undecided, the linter analyzes usage within top-level functions (in `return`
   statements or type assertions). Consistent usage can determine the type.

   ```go
   return ValueError{} // Suggests a value type

   if _, ok := err.(*PointerError); ok { /* ... */ } // Suggests a pointer type
   ```

   Note: This heuristic is a fallback and should not be relied upon for defining a type's contract.

5. **Consistent Method Receivers** (the lowest priority): As a final heuristic, if all methods on a type have a
   consistent receiver (all-value or all-pointer), that style is used.

### Designing Linter-Friendly Packages

To make an error type's intended usage explicit and ensure `errortype` can automatically determine it, add a
package-level variable assignment anywhere in the package where the error is defined:

```go
type ValueError struct{ /* ... */ }

func (v ValueError) Error() string { /* ... */ }

type PointerError struct{ /* ... */ }

func (p PointerError) Error() string { /* ... */ }

// In your package, explicitly declare the intended usage.
var (
	_ error = ValueError{}
	_ error = (*PointerError)(nil)
)
```

### Overriding Detected Types

When the linter reports ambiguous or inconsistent usage from types of an imported package that you cannot change, you
can guide the linter with an override file, see _[“Override File”](#override-file)_.

## Pointless Comparisons

Beyond error handling inconsistencies, `errortype` also detects a second class of subtle bugs: comparisons against the
address of newly created values. These comparisons, such as `ptr == &MyStruct{}` or `ptr == new(MyStruct)`, are almost
always incorrect and can lead to unexpected behavior.

### The Problem

According to the [Go language specification](https://go.dev/ref/spec#Variables), taking the address of a composite
literal (`&MyStruct{}`) or calling `new()` creates a new allocation:

> _“Calling the built-in function `new` or taking the address of a composite literal allocates storage for a variable at
> run time.”_

Each allocation gets a [unique address](https://go.dev/ref/spec#Composite_literals):

> _“Taking the address of a composite literal generates a pointer to a **unique variable** initialized with the
> literal's value.”_

This means `ptr == &MyStruct{}` will almost always evaluate to `false`, regardless of what `ptr` points to. The only
exception is zero-sized types, where [the behavior is undefined](https://go.dev/ref/spec#Address_operators):

> _“Pointers to distinct zero-size variables may or may not be equal.”_

### Examples of Problematic Code

Here are examples that `errortype` will flag:

#### Error Handling with `errors.Is`

```go
import (
	"errors"
	"log"
	"net/url"
)

func handleNetworkError(err error) {
	// This will always be false - &url.Error{} creates a unique address.
	if errors.Is(err, &url.Error{}) {
		log.Fatal("Cannot connect to service")
	}

	// Correct approach:
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		log.Fatal("Error connecting to service:", urlErr)
	}
	// ...
}
```

#### Direct Pointer Comparisons

```go
import (
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// Checking if an operator update strategy matches expected values
func validateUpdateStrategy(spec *v1alpha1.CatalogSourceSpec) {
	expectedDuration := 30 * time.Second

	// This comparison will always be false - &metav1.Duration{} creates a unique address.
	if spec.UpdateStrategy.Interval != &metav1.Duration{Duration: expectedDuration} {
		// ...
	}

	// Correct approach: Dereference the pointer and compare values (after a nil check).
	if spec.UpdateStrategy.Interval == nil || spec.UpdateStrategy.Interval.Duration != expectedDuration {
		// ...
	}
}
```

### Special Cases for `errors.Is`

`errortype` includes special handling for `errors.Is` and similar functions to reduce false positives. The linter
suppresses diagnostics when:

- **The error type has an `Unwrap() error` method**, as `errors.Is` traverses the error tree.

- **The error type has an `Is(error) bool` method**, as custom comparison logic is executed.

This behavior can be disabled with the `-check-is=false` flag.

## Override File

For more complex projects or when working with third-party libraries that have ambiguous or wrongly detected error type
usage, you may need to provide explicit guidance to the linter through an override file.

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

The `inconsistent` section is only generated by `-suggest` and is ignored by the linter. You should review all entries
and move them to the `pointer`, `value`, or `suppress` sections as appropriate.

Once your `errortypes.yaml` file is configured, use it with the `-overrides` flag:

```console
errortype -overrides=errortypes.yaml ./...
```

This instructs the linter to use your specified configuration, resolving ambiguities and suppressing noise from types
you wish to ignore.

**Note:** Always review suggestions before adding them to your override file. A suggestion makes your code consistent
with how the type is _used in your package_, but this may conflict with how the type was _designed_ to be used in its
defining package. When possible, fixing the inconsistency by refactoring the code is preferable to forcing an override.

### Overrides vs. Autodetection

It's important to understand the difference between autodetection and overrides.

- **Autodetection** runs on the package where an error type is **defined**, see
  _[“How Intended Usage is Detected”](#how-intended-usage-is-detected)_. This is the ideal place to establish the
  intended usage.

- **Overrides** are based on the usage within **your** code. They force a specific pointer or value style, overriding
  what was detected in the defining package.

**Every suggestion should be reviewed before being used as an override.** An `inconsistent` usage report may indicate a
genuine opportunity to refactor and improve your error handling. When possible, it is always better to improve detection
in the defining package by making the usage explicit, see
_[“Designing Linter-Friendly Packages”](#designing-linter-friendly-packages)_.

## Other Possible Detections

The linter currently does not look into expression `switch` logic, like

```go
	switch err {
	case io.EOF:
		return true
	}
```

A survey of more than 750 popular open source projects showed that there are nearly no issues in usages.

## Diagnostic Code Reference

`errortype` uses short diagnostic codes to categorize the issues it finds, making it easier to understand and address
specific problems.

### Error Type Consistency Issues

- **`et:ret` (Return Mismatch)**: An error type is returned incorrectly (returning a value error as a pointer or vice
  versa).

- **`et:ast` (Assertion Mismatch)**: An error type is used incorrectly in a type assertion or type switch.

- **`et:err` (Argument Mismatch)**: An error type is passed incorrectly as a target to an `errors.As`-like function.

- **`et:emb` (Embedded/Ambiguous Usage)**: The linter could not determine if an error is a pointer or value type. This
  can be resolved with an [override](#overriding-detected-types).

- **`et:var` (Variable Mismatch)**: An error type is assigned incorrectly in a
  [variable declaration](https://go.dev/ref/spec#Variable_declarations) starting with `Err` or `err`.

- **`et:rcv` (Receiver Mismatch)**: An `Unwrap` related method on a value error should be implemented with a value
  receiver, not a pointer, otherwise it wouldn't be visible.

### Pointer Comparison Issues

- **`et:cmp` (Pointless Error Comparison)**: A pointer is compared against the address of a newly created value in
  `errors.Is` or similar functions. This comparison is almost always `false`, or its result is undefined for zero-sized
  types.

  ```go
  if errors.Is(err, &url.Error{}) { // This comparison will always be false
    // ...
  }
  ```

  See also [“Pointless Comparisons”](#pointless-comparisons).

  **Fix**: Use `errors.As` to check if an error is of a certain type. If you need to check for a specific sentinel error
  instance, define it as a package-level variable and compare against that.

  ```go
  var urlErr *url.Error
  if errors.As(err, &urlErr) { // Correct approach
    // ...
  }
  ```

- **`et:equ` (Pointless Comparison)**: A pointer is compared against the address of a newly created value using equality
  operators. This comparison is almost always `false`, or its result is undefined for zero-sized types.

  ```go
  if ptr == &MyStruct{} { // This comparison will always be false
    // ...
  }
  ```

  **Fix**: Dereference the pointer and compare the underlying values (`*ptr == MyStruct{}`). If you need to check for a
  specific sentinel instance, define it as a package-level variable and compare against that.

### Other and Style Issues

- **`et:unw` (Calling Unwrap)**: An unwrapping function (like `errors.Is`) was found inside an `Is(error) bool` method.

  ```go
  func (MyEOFError) Is(err error) bool {
    return errors.Is(err, io.ErrUnexpectedEOF)
  }
  ```

  See [function `Is(err, target)`](https://pkg.go.dev/errors#Is):

  > _“An `Is` method should only shallowly compare `err` and the `target` and not call `Unwrap` on either.”_

  **Fix**: Replace the call with a direct, shallow comparison, such as `err == io.ErrUnexpectedEOF`.

  ```go
  func (MyEOFError) Is(err error) bool {
    return err == io.ErrUnexpectedEOF
  }
  ```

- **`et:sty` (Style Mismatch)**: The target argument to an `errors.As`-like function is not an address operation on a
  variable.

  ```go
    // Confusing: checks for a value, but looks like a pointer check.
    if errors.As(err, &MyError{}) { /* ... */ }
  ```

  While this is valid code, it can be misleading. The expression `&MyError{}` creates a pointer to a struct literal.
  When passed to `errors.As`, it checks if `err` wraps a `MyError` _value_, not a pointer. This is easily misread,
  obfuscating the check's true intent.

  **Fix**: To improve clarity, declare a variable of the target error type and pass its address to `errors.As`.

  ```go
    var e MyError // Or "var e *MyError" for pointer errors
    if errors.As(err, &e) { /* ... */ } // The clear, recommended style
  ```

  This longer form is unambiguous and clearly states the type being checked for.

- **`et:arg` (Invalid Argument)**: The target argument to an `errors.As`-like function is invalid (not a pointer to a
  type implementing `error`). This is also flagged by the standard Go
  [`errorsas`](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/errorsas) linter.

- **`et:sig` (Wrong Signature)**: An `Unwrap` related method has the wrong signature. This is also flagged by the
  standard Go [`stdmethods`](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/stdmethods) linter.

- **`et:unu` (Unused Result)**: The result of an `errors.Is`-like function is unused.

  ```go
    errors.Is(err, &MyError{})
  ```

- **`et:uca` (Unchecked Type Assert)**: An unchecked type assert might lead to a run-time panic on a wrapped error.

  ```go
    if err.(*net.AddrError).Err == "missing port in address" { /* ... */ }
  ```

  **Fix**: While sometimes valid, prefer `errors.As`.

## Integration

### `golangci-lint` Module Plugin

Add a file `.custom-gcl.yaml` to your source with

```yaml
---
version: v2.6.2

name: golangci-lint
destination: .

plugins:
  - module: fillmore-labs.com/errortype
    import: fillmore-labs.com/errortype/gclplugin
    version: v0.0.8
```

then run `golangci-lint custom` from your project root. You get a custom `golangci-lint` executable that can be
configured in `.golangci.yaml`:

```YAML
---
version: "2"
linters:
  enable:
    - errortype
  settings:
    custom:
      errortype:
        type: module
        description: "errortype helps prevent subtle bugs in error handling."
        original-url: "https://fillmore-labs.com/errortype"
        settings:
          overrides:
            pointer:
              - test/a.PointerOverride
            value:
              - test/a.ValueOverride
            suppress:
              - test/a.SuppressOverride
          style-check: true
          deep-is-check: false
          check-is: true
          unchecked-assert: false
          check-unused: false
```

and can be used like `golangci-lint`:

```shell
./golangci-lint run .
```

See also the golangci-lint
[module plugin system documentation](https://golangci-lint.run/plugins/module-plugins/#the-automatic-way).

## Links

- [Background on the problem this linter solves](https://blog.fillmore-labs.com/posts/errors-1/)
- [Why you shouldn't call `Unwrap` in `Is(error) bool` methods](https://blog.fillmore-labs.com/posts/errors-2/)
- [Enhanced error inspection with better ergonomics](https://pkg.go.dev/fillmore-labs.com/exp/errors)
- [Proposal for a generic `errors.As` variant](http://go.dev/issues/51945)

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.
