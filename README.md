# Errortype

[![Go Reference](https://pkg.go.dev/badge/fillmore-labs.com/errortype.svg)](https://pkg.go.dev/fillmore-labs.com/errortype)
[![Test](https://github.com/fillmore-labs/errortype/actions/workflows/test.yaml/badge.svg?branch=dev)](https://github.com/fillmore-labs/errortype/actions/workflows/test.yaml?query=branch%3Adev)
[![CodeQL](https://github.com/fillmore-labs/errortype/actions/workflows/github-code-scanning/codeql/badge.svg?branch=dev)](https://github.com/fillmore-labs/errortype/actions/workflows/github-code-scanning/codeql?query=branch%3Adev)
[![Coverage](https://codecov.io/gh/fillmore-labs/errortype/branch/dev/graph/badge.svg?token=MMLHL14ZP6)](https://codecov.io/gh/fillmore-labs/errortype)
[![Go Report Card](https://goreportcard.com/badge/fillmore-labs.com/errortype)](https://goreportcard.com/report/fillmore-labs.com/errortype)
[![Codeberg CI](https://ci.codeberg.org/api/badges/15305/status.svg?branch=dev)](https://ci.codeberg.org/repos/15305/branches/dev)
[![License](https://img.shields.io/badge/license-Apache--2.0-green)](https://www.apache.org/licenses/LICENSE-2.0)

`errortype` is an analysis tool that performs static checks across three main categories:

1. **Inconsistent Error Type Usage**: Ensures error types are used consistently as either pointers or values in returns,
   type assertions, `errors.AsType`/`errors.As` calls, and wrapped with `%w` in `fmt.Errorf`.

2. **Pointless Comparisons**: Detects comparisons against newly allocated addresses (like `ptr == &MyStruct{}` or
   `errors.Is(err, &url.Error{})`), which are almost always incorrect.

3. **Error Naming Conventions** _(opt-in)_: Checks that sentinel error variables use the `Err` prefix (e.g.,
   `ErrNotFound`) and structured error types use the `Error` suffix (e.g., `ParseError`), following the
   [Go naming conventions](#error-naming-conventions).

## Getting Started

### Installation

#### Go

```shell
go install fillmore-labs.com/errortype@latest
```

#### Homebrew

```shell
brew install fillmore-labs/tap/errortype
```

#### Eget

[Install `eget`](https://github.com/zyedidia/eget#how-to-get-eget), then:

```shell
eget fillmore-labs/errortype
```

## Usage

Analyze your entire project:

```shell
errortype -recommended ./...
```

### Command-Line Flags

| Flag                  | Description                                                                                                                                     | Default               |
| --------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- | --------------------- |
| `-recommended`        | Enables all recommended checks. Unstable between releases and not suitable for CI use, but recommended for development in conformant code bases. | `false`               |
| `-non-comparable`     | Extended checks for [not comparable](https://go.dev/ref/spec#Comparison_operators) error types _(recommended)_                                  | `false`               |
| `-legacy`             | Check for [pre Go 1.13](https://go.dev/blog/go1.13-errors#examining-errors-with-is-and-as) error query helpers _(recommended)_                  | `false`               |
| `-naming`             | Check error [naming conventions](#error-naming-conventions) _(recommended)_                                                                     | `false`               |
| `-style-check`        | Check for confusing uses of `errors.As` _(recommended)_                                                                                         | `false`               |
| `-check-unused`       | Report unused results of `errors.Is/As/AsType`-like functions                                                                                   | `true`                |
| `-check-is`           | Suppress diagnostics on `errors.Is` if the type has an `Is(error) bool` or `Unwrap() error` method                                              | `true`                |
| `-deep-is-check`      | In `Is` methods, diagnose any unwrapping call, not just those using `target`                                                                    | `false`               |
| `-unchecked-assert`   | Diagnose unchecked type asserts on errors                                                                                                       | `false`               |
| `-prefix-filter`      | Restrict variable analysis to explicitly named variables (`err` or `Err`)                                                                       | `true`                |
| `-c <N>`              | Lines of context around each issue (`-1` = none, `0` = offending line only)                                                                     | `-1`                  |
| `-test`               | Analyze test files                                                                                                                              | `true`                |
| `-fix`                | Apply suggested fixes (e.g., [naming](#error-naming-conventions) renames, [legacy](#legacy-assertion-query) rewrites)                           | `false`               |
| `-overrides <file>`   | Read type overrides from a YAML file (see [Override File](#override-file))                                                                      | -                     |
| `-suggest <file>`     | Append suggestions to an override file (`-` for stdout)                                                                                         | -                     |
| `-heuristics <list>`  | Heuristics to use (`"off"` to disable) _(Experimental)_                                                                                         | `var,usage,receivers` |
| `-tracetypes <regex>` | Trace type detection in matching packages _(Experimental)_                                                                                      | -                     |

## Inconsistent Error Type Usage

A common and subtle bug occurs when error types are used inconsistently: sometimes as values, sometimes as pointers.
This can cause `errors.As` checks to silently fail.

For example, this code ([Go Playground](https://go.dev/play/p/kvaL1A-Pkk9)):

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

	if keySizeErr, ok := errors.AsType[*aes.KeySizeError](err); ok {
		fmt.Printf("AES keys must be 16, 24, or 32 bytes long, got %d bytes.\n", keySizeErr)
	} else if err != nil {
		fmt.Println(err)
	}
}
```

It prints a generic error message because [`aes.KeySizeError`](https://pkg.go.dev/crypto/aes#KeySizeError) is a value
type. Running `errortype .` reports:

```console
.../main.go:13:23: Error type "crypto/aes.KeySizeError" should be queried ⏎
    as a value ("errors.AsType[aes.KeySizeError]"), not a pointer. (et:ast)
```

Changing line 13 to `... := errors.AsType[aes.KeySizeError](err); ...` fixes this.

### How Intended Usage Is Detected

The linter determines an error type's intended use (pointer vs. value) by analyzing its _defining_ package, in order of
precedence:

1. **Overrides**: User-defined overrides (see [Override File](#override-file)) take highest priority.

2. **`Unwrap`-related methods**: Methods like `Is`, `As`, and `Unwrap` with pointer receivers are only visible when the
   error is used as a pointer.

   ```go
   func (e *PointerError) Unwrap() error { /* ... */ } // Only visible when the dynamic value in the error interface is a pointer (&PointerError{}).
   ```

3. **Package-level variable assignments**: `var _ error = ...` declarations explicitly state intent.

   ```go
   var _ error = ValueError{}         // Asserts ValueError should be used as a value type.
   var _ error = (*PointerError)(nil) // Asserts PointerError should be used as a pointer type.
   ```

4. **Usage in functions**: Consistent usage in `return` statements or type assertions.

   ```go
   return ValueError{} // Suggests value type

   if _, ok := err.(*PointerError); ok { /* ... */ } // Suggests pointer type
   ```

5. **Consistent method receivers**: If all methods have the same receiver type, that style is used.

   > [!NOTE]
   >
   > This heuristic is a fallback and should not be relied upon for defining a type's contract.

### Designing Linter-Friendly Packages

To make intent explicit, add a variable assignment in the _defining_ package:

```go
type ValueError struct { /* ... */ }

func (v ValueError) Error() string { /* ... */ }

type PointerError struct { /* ... */ }

// The value receiver here is intentional: either the API
// can't change, or it's a deliberate choice for this type.
func (p PointerError) Error() string { /* ... */ }

var (
	// Explicitly declare intended usage.
	_ error = ValueError{}
	// Despite the value receiver, the var declaration clarifies the intended pointer usage.
	_ error = (*PointerError)(nil)
)
```

### Overriding Detected Types

When the linter reports ambiguous usage from an imported package that you cannot modify, use an override file (see
[Override File](#override-file)).

## Pointless Comparisons

`errortype` detects comparisons against newly allocated addresses. Per the [Go spec](https://go.dev/ref/spec#Variables),
`&MyStruct{}` and `new(T)` each create a unique address, so `ptr == &MyStruct{}` is almost always `false`. For
zero-sized types, [the result is undefined](https://go.dev/ref/spec#Comparison_operators).

### Examples

#### Error Handling with `errors.Is`

```go
import (
	"errors"
	"log"
	"net/url"
)

func handleNetworkError(err error) {
	// Always false, &url.Error{} creates a unique address.
	if errors.Is(err, &url.Error{}) {
		log.Fatal("Cannot connect to service")
	}

	// Correct approach:
	if err, ok := errors.AsType[*url.Error](err); ok {
		log.Fatal("Error connecting to service:", err.URL)
	}
	// ...
}
```

#### Direct Pointer Comparisons

```go
import (
	"time"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updatedInterval(catsrc *v1alpha1.CatalogSource) {
	expectedTime := 30 * time.Second

	// Always true: &metav1.Duration{} creates a unique address.
	if (catsrc.Spec.UpdateStrategy.Interval != &metav1.Duration{Duration: expectedTime}) {
		// ...
	}

	// Correct: compare values.
	if catsrc.Spec.UpdateStrategy.Interval == nil || catsrc.Spec.UpdateStrategy.Interval.Duration != expectedTime {
		// ...
	}
}
```

### Special Cases for `errors.Is`

The linter suppresses diagnostics when the error type has an `Is(error) bool` or `Unwrap() error` method (since
`errors.Is` traverses the chain, and an error could have custom comparison logic). Disable with `-check-is=false`.

## Wrapper Functions

`errortype` automatically detects wrapper functions around [`errors.Is`](https://pkg.go.dev/errors#Is),
[`errors.AsType`](https://pkg.go.dev/errors#AsType), [`errors.As`](https://pkg.go.dev/errors#As), and
[`fmt.Errorf`](https://pkg.go.dev/fmt#Errorf) within the analyzed package, and validates their call-site arguments the
same way it does for the standard library functions.

For example, a test helper like this is recognized as an `errors.As` wrapper:

```go
func RequireErrorsAs(t *testing.T, err error, target any, format string, args ...any) {
	t.Helper()

	if !errors.As(err, target) {
		t.Fatalf(format, args...)
	}
}
```

Because `err` and `target` are forwarded to `errors.As`, call sites of `RequireErrorsAs` receive the same checks.

## Error Assertion Queries

`errortype` detects and flags “error assertion queries”: functions or methods that check if an error matches a type
using a simple type assertion:

```go
type MyError struct{ msg string }

func (e *MyError) Error() string { return e.msg }

func (e *MyError) Is(err error) bool {
	_, ok := err.(*MyError) // Is-method assertion query (et:ias)
	return ok
}

func IsMyError(err error) bool {
	_, ok := err.(*MyError) // Legacy assertion query (et:lgc)
	return ok
}
```

This pattern is problematic depending on its context:

### Legacy Assertion Query

Standalone error query functions like `func IsMyError(err error) bool` that rely on a type assertion are a legacy
pattern from before Go 1.13. They fail to account for wrapped errors (e.g., created by `fmt.Errorf("...: %w", err)`).

```go
	err1 := &MyError{"error 1"}
	wrapped := fmt.Errorf("wrapped: %w", err1)
	if IsMyError(wrapped) {
		// should be true
	}
```

When the `-legacy` flag is enabled, `errortype` flags these functions and suggests replacing them with `errors.AsType`
or `errors.As` checks to properly query the error chain.

### `Is`-Method Assertion Query

Implementing an `Is(err error) bool` method on an error type that simply asserts the type of the target breaks the
semantics of `errors.Is`. The `Is` method is intended for equivalence with existing errors, but a type assertion turns
it into a type check (already covered by `errors.As`). This causes `errors.Is` to incorrectly match any error value of
the same type, rather than matching specific instances or logical equivalents.

```go
	err1 := &MyError{"error 1"}
	err2 := &MyError{"error 2"}
	wrapped := fmt.Errorf("wrapped: %w", err1)
	if errors.Is(wrapped, err2) {
		// should not be true
	}
```

This is an active misuse of the Go 1.13 error API and is flagged unconditionally. You should remove the `Is` method and
use `errors.AsType` when you need to match by type.

## Error Naming Conventions

When enabled, `errortype` checks that error names follow [Go conventions](https://go.dev/wiki/Errors#naming):

- **Sentinel errors** (package-level variables) should start with `Err` (or `err` for unexported), e.g., `ErrNotFound`.
- **Structured error types** should end with `Error`, e.g., `ParseError`.

This flag is **off by default** since it can be noisy on existing codebases, but it is recommended for new projects and
those that already follow these rules.

The `-fix` flag automatically renames most non-compliant declarations. When it renames an exported declaration, it keeps
the old name with a [`// Deprecated:`](https://go.dev/wiki/Deprecated) comment for backward compatibility. Dependent
packages will continue to compile and can later be
[updated automatically](https://go.dev/blog/inliner#source-level-inlining) via `go fix -inline ./...`.

You can also run this check as a standalone tool (`errorname`). Install it with
`go install fillmore-labs.com/errortype/errorname@latest`, then apply fixes with
`go fix -fixtool=$(which errorname) ./...`.

### Why Adhere to Error Naming Conventions

In Go, names are supposed to communicate intent. The `Err` prefix signals _“this is an error sentinel”_ which can be
used with `errors.Is` and the likewise `Error` suffix a structured error type that can be queried with `errors.AsType`.
Also, static analyzers (including `errortype`'s own `-prefix-filter`) can rely on these naming conventions to accurately
scope their analysis.

## Override File

For third-party libraries with ambiguous error types, provide an override file.

Generate a sample with:

```shell
errortype -suggest=errortypes.yaml ./...
```

This creates a file with the following structure:

```yaml
# Override types for your.path/package
---
pointer: # Types that should always be used as pointers
  - imported.path/one.PointerOverride

value: # Types that should always be used as values
  - imported.path/two.ValueOverride

suppress: # Types to ignore during analysis
  - imported.path/one.ErrorToIgnore

inconsistent: # Types with inconsistent usage (generated by -suggest, ignored by linter)
  - imported.path/two.InconsistentUsage
```

Review entries in `inconsistent` and move them to `pointer`, `value`, or `suppress` as appropriate. Then run:

```shell
errortype -overrides=errortypes.yaml ./...
```

> [!NOTE]
>
> A suggestion makes your code consistent with how the type is _used in your package_, but this may conflict with its
> intended design. Refactoring is often preferable to overriding.

### Overrides vs. Autodetection

- **Autodetection** runs on the package where an error type is _defined_ (see
  [How Intended Usage Is Detected](#how-intended-usage-is-detected)).
- **Overrides** force a style based on usage in _your_ code, overriding autodetection.

When possible, improve detection in the defining package by making usage explicit (see
[Designing Linter-Friendly Packages](#designing-linter-friendly-packages)).

## Diagnostic Code Reference

`errortype` uses short codes to categorize issues. These are experimental, intended for filtering and aggregation rather
than as stable cross-version identifiers:

| Code     | Name                        | Description                                                                                                                                          |
| -------- | --------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| `et:arg` | Invalid Argument            | Invalid target to an `errors.As`-like function, **possible panic**.                                                                                  |
| `et:asn` | Assignment Mismatch         | Incorrect type in assignment.                                                                                                                        |
| `et:ast` | Assertion Mismatch          | Incorrect type in a type assertion, type switch, or `errors.As/AsType` call.                                                                         |
| `et:cmp` | Pointless Error Comparison  | Comparison against `&T{}` in `errors.Is`: always false. Use `errors.As/AsType` instead.                                                              |
| `et:emb` | Ambiguous Usage             | Could not determine if error is pointer or value type; use an [override](#overriding-detected-types).                                                |
| `et:equ` | Pointless Comparison        | Pointer compared against `&T{}`: always false (undefined for zero-sized types).                                                                      |
| `et:err` | Argument Mismatch           | Incorrect target passed to an `errors.As`-like function.                                                                                             |
| `et:ias` | `Is`-Method Assertion Query | `Is` method implements an error assertion query.                                                                                                     |
| `et:lgc` | Legacy Assertion Query      | Legacy error assertion query does not account for wrapped errors.                                                                                    |
| `et:nam` | Error Naming                | Error variables should start with `Err` or `err`; structured error type names should end with `Error`.                                               |
| `et:nce` | Non-comparable Value Type   | A [not comparable](https://go.dev/ref/spec#Comparison_operators) value type should have an `Is` method.                                              |
| `et:rcv` | Receiver Mismatch           | `Unwrap`-related method on a value error should use value receiver.                                                                                  |
| `et:ret` | Return Mismatch             | Error type returned incorrectly (value as pointer or vice versa).                                                                                    |
| `et:sig` | Wrong Signature             | `Unwrap`-related method has a wrong signature (also flagged by [`stdmethods`](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/stdmethods)). |
| `et:sty` | Style Mismatch              | Target to `errors.As` is not an address operation; declare a variable for clarity.                                                                   |
| `et:uca` | Unchecked Type Assert       | Unchecked type assert might panic on wrapped error; prefer `errors.As/AsType`.                                                                       |
| `et:unu` | Unused Result               | Result of `errors.Is/As/AsType`, `fmt.Errorf` (or wrapper) is unused.                                                                                |
| `et:unw` | Calling Unwrap              | Unwrapping function called inside `Is(error) bool`; use shallow comparison (direct type assertion) instead.                                          |
| `et:var` | Variable Mismatch           | Incorrect type in a variable declaration; only `Err/err`-prefixed variables are checked when `-prefix-filter` is on (default).                       |
| `et:wrp` | Wrap Mismatch               | Error wrapped as a pointer instead of a value (or vice versa) by `%w` in `fmt.Errorf`.                                                               |

## Integration

### `golangci-lint` Module Plugin

Add `.custom-gcl.yaml` to your project:

```yaml
---
version: v2.12.2

name: golangci-lint
destination: .

plugins:
  - module: fillmore-labs.com/errortype
    import: fillmore-labs.com/errortype/gclplugin
    version: v0.0.12
```

Run `golangci-lint custom` to build a custom executable. Configure in `.golangci.yaml` (the values under `settings:`
below differ from the defaults; adapt them to your project):

```yaml
---
version: "2"
linters:
  default: none
  enable:
    - errortype
  settings:
    custom:
      errortype:
        type: module
        description: errortype helps prevent subtle bugs in error handling.
        original-url: https://fillmore-labs.com/errortype
        settings:
          recommended: true
          naming: true
          legacy: true
          non-comparable: true
          style-check: true
          deep-is-check: false
          check-is: true
          unchecked-assert: false
          check-unused: false
          prefix-filter: false
          overrides:
            pointer:
              - your.pkg/a.PointerOverrideError
            value:
              - your.pkg/a.ValueOverrideError
            suppress:
              - your.pkg/a.SuppressOverrideError
```

Then run:

```shell
./golangci-lint run .
```

See the [module plugin documentation](https://golangci-lint.run/plugins/module-plugins/#the-automatic-way).

## Links

- [Background on the problem this linter solves](https://blog.fillmore-labs.com/posts/errors-1/)
- [Why you shouldn't call `Unwrap` in `Is(error) bool` methods](https://blog.fillmore-labs.com/posts/errors-2/)
- [Go Wiki: Error naming](https://go.dev/wiki/Errors#naming)

## License

Licensed under the Apache License 2.0. See [LICENSE](LICENSE) for details.
