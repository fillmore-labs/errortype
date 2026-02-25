# Errortype

[![Go Reference](https://pkg.go.dev/badge/fillmore-labs.com/errortype.svg)](https://pkg.go.dev/fillmore-labs.com/errortype)
[![Test](https://github.com/fillmore-labs/errortype/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/fillmore-labs/errortype/actions/workflows/test.yml?query=branch%3Amain)
[![CodeQL](https://github.com/fillmore-labs/errortype/actions/workflows/github-code-scanning/codeql/badge.svg?branch=main)](https://github.com/fillmore-labs/errortype/actions/workflows/github-code-scanning/codeql?query=branch%3Amain)
[![Coverage](https://codecov.io/gh/fillmore-labs/errortype/branch/main/graph/badge.svg?token=MMLHL14ZP6)](https://codecov.io/gh/fillmore-labs/errortype)
[![Go Report Card](https://goreportcard.com/badge/fillmore-labs.com/errortype)](https://goreportcard.com/report/fillmore-labs.com/errortype)
[![Codeberg CI](https://ci.codeberg.org/api/badges/15305/status.svg?branch=main)](https://ci.codeberg.org/repos/15305/branches/main)
[![License](https://img.shields.io/github/license/fillmore-labs/errortype)](https://www.apache.org/licenses/LICENSE-2.0)

`errortype` is a static analysis tool that performs three checks:

1. **Inconsistent Error Type Usage**: Ensures error types are used consistently as either pointers or values in returns,
   type assertions, and `errors.As` calls.

2. **Pointless Comparisons**: Detects comparisons against newly allocated addresses (like `errors.Is(err, &url.Error{})`
   or `ptr == &MyStruct{}`), which are almost always incorrect.

3. **Error Naming Conventions** _(opt-in)_: Checks that sentinel error variables use the `Err` prefix (e.g.,
   `ErrNotFound`) and structured error types use the `Error` suffix (e.g., `ParseError`), following the
   [Go naming conventions](https://go.dev/wiki/Errors#naming).

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
errortype -naming -style-check ./...
```

### Command-Line Flags

| Flag                  | Description                                                                    | Default               |
| --------------------- | ------------------------------------------------------------------------------ | --------------------- |
| `-naming`             | Check error [naming conventions](#error-naming-conventions) (recommended)      | `false`               |
| `-style-check`        | Check for confusing uses of `errors.As` (recommended)                          | `false`               |
| `-check-unused`       | Report unused results of `errors.As`-like functions                            | `true`                |
| `-check-is`           | Suppress diagnostics on `errors.Is` if the type has an `Is(error) bool` method | `true`                |
| `-deep-is-check`      | In `Is` methods, diagnose any unwrapping call, not just those using `target`   | `false`               |
| `-unchecked-assert`   | Diagnose unchecked type asserts on errors                                      | `false`               |
| `-c <N>`              | Lines of context around each issue (`-1` = none, `0` = offending line only)    | `-1`                  |
| `-test`               | Analyze test files                                                             | `true`                |
| `-overrides <file>`   | Read type overrides from a YAML file (see [Override File](#override-file))     | —                     |
| `-suggest <file>`     | Append suggestions to an override file (`-` for stdout)                        | —                     |
| `-heuristics <list>`  | Heuristics to use (`“off”` to disable) _(Experimental)_                        | `var,usage,receivers` |
| `-tracetypes <regex>` | Trace type detection in matching packages _(Experimental)_                     | —                     |

## Inconsistent Error Type Usage

A common and subtle bug occurs when error types are used inconsistently — sometimes as values, sometimes as pointers.
This can cause `errors.As` checks to silently fail.

Consider this code ([Go Playground](https://go.dev/play/p/m4SEPqkZ2Zu)):

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

This prints the generic error because `aes.KeySizeError` is a value type, not a pointer. Changing line 13 to
`var kse aes.KeySizeError` fixes it.

Running `errortype .` reports:

```console
.../main.go:14:20: Target for value error "crypto/aes.KeySizeError" ⏎
    is a pointer-to-pointer, use a pointer to a value instead: ⏎
    "var kse aes.KeySizeError; ... errors.As(err, &kse)". (et:err)
```

### How Intended Usage Is Detected

The linter determines an error type's intended use (pointer vs. value) by analyzing its _defining_ package, in order of
precedence:

1. **Overrides**: User-defined overrides (see [Override File](#override-file)) take highest priority.

2. **`Unwrap` related methods**: Methods like `Is`, `As`, and `Unwrap` with pointer receivers are only visible when the
   error is used as a pointer.

   ```go
   func (e *PointerError) Unwrap() error { /* ... */ } // Only visible from error(&PointerError{}).
   ```

3. **Package-level variable assignments**: `var _ error = ...` declarations explicitly state intent.

   ```go
   var _ error = ValueError{}         // Declares ValueError as a value type.
   var _ error = (*PointerError)(nil) // Declares PointerError as a pointer type.
   ```

4. **Usage in functions**: Consistent usage in `return` statements or type assertions.

   ```go
   return ValueError{} // Suggests value type

   if _, ok := err.(*PointerError); ok { /* ... */ } // Suggests pointer type
   ```

   > [!NOTE]
   >
   > This heuristic is a fallback and should not be relied upon for defining a type's contract.

5. **Consistent method receivers**: If all methods have the same receiver type, that style is used.

### Designing Linter-Friendly Packages

To make intent explicit, add a variable assignment in the declaring package:

```go
type ValueError struct{ /* ... */ }

func (v ValueError) Error() string { /* ... */ }

type PointerError struct{ /* ... */ }

func (p PointerError) Error() string { /* ... */ }

// Explicitly declare intended usage.
var (
	_ error = ValueError{}
	_ error = (*PointerError)(nil)
)
```

### Overriding Detected Types

When the linter reports ambiguous usage from an imported package you cannot modify, use an override file (see
[Override File](#override-file)).

## Wrapper Functions

`errortype` automatically detects wrapper functions around [`errors.Is`](https://pkg.go.dev/errors#Is),
[`errors.As`](https://pkg.go.dev/errors#As), and [`errors.AsType`](https://pkg.go.dev/errors#AsType) within the analyzed
package. This allows the linter to validate arguments passed to your custom error helpers just as it does for the
standard library functions.

For example, a test helper like this is recognized as an `errors.As` wrapper:

```go
func RequireErrorsAs(t *testing.T, err error, target any, format string, args ...any) {
	t.Helper()

	if !errors.As(err, target) {
		t.Fatalf(format, args...)
	}
}
```

The linter detects that `err` and `target` are forwarded to `errors.As` and applies the same checks to call sites of
`RequireErrorsAs`.

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
	// Always false — &url.Error{} creates a unique address.
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

func validateUpdateStrategy(spec *v1alpha1.CatalogSourceSpec) {
	expectedDuration := 30 * time.Second

	// Always false — &metav1.Duration{} creates a unique address.
	if spec.UpdateStrategy.Interval != &metav1.Duration{Duration: expectedDuration} {
		// ...
	}

	// Correct: compare values after nil check.
	if spec.UpdateStrategy.Interval == nil || spec.UpdateStrategy.Interval.Duration != expectedDuration {
		// ...
	}
}
```

### Special Cases for `errors.Is`

The linter suppresses diagnostics when the error type has an `Unwrap() error` method (since `errors.Is` traverses the
chain) or an `Is(error) bool` method (custom comparison logic). Disable with `-check-is=false`.

## Error Naming Conventions

When enabled with `-naming`, `errortype` checks that error names follow
[Go conventions](https://go.dev/wiki/Errors#naming):

- **Sentinel errors** (variables) should start with `Err` (or `err` for unexported), e.g., `ErrNotFound`.
- **Structured error types** should end with `Error`, e.g., `ParseError`.

This flag is **off by default** since it can be noisy on existing codebases, but its usage is recommended for new
projects and codebases that already follow these conventions.

See also
[Go: Naming Errors and the Parlance of Error Types](https://www.matttproud.com/blog/posts/go-error-parlance.html) for a
detailed discussion.

## Override File

For complex projects or third-party libraries with ambiguous error types, provide an override file.

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

`errortype` uses short codes to categorize issues:

### Error Type Consistency

| Code     | Name               | Description                                                                                           |
| -------- | ------------------ | ----------------------------------------------------------------------------------------------------- |
| `et:ret` | Return Mismatch    | Error type returned incorrectly (value as pointer or vice versa)                                      |
| `et:ast` | Assertion Mismatch | Incorrect type in assertion or switch                                                                 |
| `et:err` | Argument Mismatch  | Incorrect target passed to `errors.As`-like function                                                  |
| `et:emb` | Ambiguous Usage    | Could not determine if error is pointer or value type — use an [override](#overriding-detected-types) |
| `et:var` | Variable Mismatch  | Incorrect assignment in variable declaration starting with `Err`/`err`                                |
| `et:rcv` | Receiver Mismatch  | `Unwrap`-related method on value error should use value receiver                                      |

### Pointer Comparisons

| Code     | Name                       | Description                                                                       |
| -------- | -------------------------- | --------------------------------------------------------------------------------- |
| `et:cmp` | Pointless Error Comparison | Comparison against `&T{}` in `errors.Is` — always false. Use `errors.As` instead. |
| `et:equ` | Pointless Comparison       | Pointer compared against `&T{}` — always false. Dereference and compare values.   |

### Naming Conventions

| Code      | Name            | Description                                                      |
| --------- | --------------- | ---------------------------------------------------------------- |
| `et:nam`  | Variable Naming | Sentinel error name should start with `Err` (exported) or `err`. |
| `et:nam+` | Type Naming     | Structured error type name should end with `Error`.              |

### Other Issues

| Code     | Name                  | Description                                                                                                                                        |
| -------- | --------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| `et:unw` | Calling Unwrap        | Unwrapping function called inside `Is(error) bool` — use shallow comparison instead.                                                               |
| `et:sty` | Style Mismatch        | Target to `errors.As` is not an address operation — declare a variable for clarity.                                                                |
| `et:arg` | Invalid Argument      | Invalid target to an `errors.As`-like function, **possible panic**.                                                                                |
| `et:sig` | Wrong Signature       | `Unwrap`-related method has wrong signature (also flagged by [`stdmethods`](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/stdmethods)). |
| `et:unu` | Unused Result         | Result of `errors.Is`-like function is unused.                                                                                                     |
| `et:uca` | Unchecked Type Assert | Unchecked type assert might panic on wrapped error — prefer `errors.As`.                                                                           |

## Integration

### `golangci-lint` Module Plugin

Add `.custom-gcl.yaml` to your project:

```yaml
---
version: v2.10.1

name: golangci-lint
destination: .

plugins:
  - module: fillmore-labs.com/errortype
    import: fillmore-labs.com/errortype/gclplugin
    version: v0.0.10
```

Run `golangci-lint custom` to build a custom executable. Configure in `.golangci.yaml`:

```yaml
---
version: "2"
linters:
  enable:
    - errortype
  settings:
    custom:
      errortype:
        type: module
        description: errortype helps prevent subtle bugs in error handling.
        original-url: https://fillmore-labs.com/errortype
        settings:
          naming: true
          style-check: true
          deep-is-check: false
          check-is: true
          unchecked-assert: false
          check-unused: false
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
