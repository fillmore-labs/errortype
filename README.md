# errortype

[![Go Reference](https://pkg.go.dev/badge/fillmore-labs.com/errortype.svg)](https://pkg.go.dev/fillmore-labs.com/errortype)
[![Test](https://github.com/fillmore-labs/errortype/actions/workflows/test.yaml/badge.svg?branch=main)](https://github.com/fillmore-labs/errortype/actions/workflows/test.yaml?query=branch%3Amain)
[![CodeQL](https://github.com/fillmore-labs/errortype/actions/workflows/github-code-scanning/codeql/badge.svg?branch=main)](https://github.com/fillmore-labs/errortype/actions/workflows/github-code-scanning/codeql?query=branch%3Amain)
[![Coverage](https://codecov.io/gh/fillmore-labs/errortype/branch/main/graph/badge.svg?token=MMLHL14ZP6)](https://codecov.io/gh/fillmore-labs/errortype)
[![Codeberg CI](https://ci.codeberg.org/api/badges/15305/status.svg?branch=main)](https://ci.codeberg.org/repos/15305/branches/main)
[![License](https://img.shields.io/badge/license-Apache--2.0-green)](https://www.apache.org/licenses/LICENSE-2.0)

`errortype` is a static analysis tool with four kinds of checks:

1. **[Inconsistent Error Type Usage](#inconsistent-error-type-usage)**: Ensures each error type is used consistently,
   either as a pointer or as a value.

2. **[Pointless Comparisons](#pointless-comparisons)**: Detects comparisons against newly allocated addresses (like
   `ptr == &MyStruct{}` or `errors.Is(err, &url.Error{})`), which are almost always incorrect.

3. **[Error Assertion Queries](#error-assertion-queries)**: Flags `Is` methods that are really type checks, and
   _(opt-in)_ pre-Go 1.13 helpers like `IsNotFound(err)` that ignore wrapped errors.

4. **[Error Naming Conventions](#error-naming-conventions)** _(opt-in)_: Checks that sentinel error variables use the
   `Err` prefix, like `ErrNotFound`, and structured error types use the `Error` suffix, like `ParseError`. Most findings
   can be fixed automatically with `-fix`.

If you use `gopls` v0.24.0 or later, you already have its `ptrtoerror` analyzer, which checks conversions to `error` as
you write code and asks you to declare new error types with the same `var _ error = ...` assertion `errortype`
[recommends](#designing-linter-friendly-packages). `errortype` uses these declarations but doesn't need them. It runs
before you commit or in CI. See [Related Tools](#related-tools).

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

Analyze your entire project during development:

```shell
errortype -recommended ./...
```

`-recommended` additionally enables the opt-in checks `-naming`, `-legacy`, `-non-comparable`, `-style-check`, and
`-receivers`. The set may vary between releases, so in CI enable the individual checks your codebase already passes
instead, for example:

```shell
errortype -naming ./...
```

### Command-Line Flags

| Flag                  | Description                                                                                                                                                         | Default |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| `-recommended`        | Enables all recommended opt-in checks, meant for development (see [Usage](#usage))                                                                                  | `false` |
| `-non-comparable`     | Extended checks for error types that are [not comparable](https://go.dev/ref/spec#Comparison_operators) _(recommended)_                                             | `false` |
| `-legacy`             | Check for [pre-Go 1.13](https://go.dev/blog/go1.13-errors#examining-errors-with-is-and-as) error query helpers _(recommended)_                                      | `false` |
| `-naming`             | Check error [naming conventions](#error-naming-conventions) _(recommended)_                                                                                         | `false` |
| `-receivers`          | Check for a pointer receiver on `Error` for [pointer errors](#how-intended-usage-is-detected), value receivers on exported methods for value errors _(recommended)_ | `false` |
| `-style-check`        | Check for confusing uses of `errors.As` _(recommended)_                                                                                                             | `false` |
| `-check-unused`       | Report unused results of `errors.Is/As/AsType`, `fmt.Errorf`, and their wrappers                                                                                    | `true`  |
| `-strict-is-target`   | Report pointless comparisons on `errors.Is` also in [special cases](#special-cases-for-errorsis)                                                                    | `false` |
| `-strict-is-method`   | In `Is` methods, diagnose any unwrapping call, not just those using `target`                                                                                        | `false` |
| `-unchecked-assert`   | Diagnose unchecked type asserts on errors                                                                                                                           | `false` |
| `-prefix-filter`      | Restrict variable analysis to variables named like errors (`err…`, `…Err`, exported `Err…`)                                                                         | `true`  |
| `-c <N>`              | Lines of context around each issue (`-1` = none, `0` = offending line only)                                                                                         | `-1`    |
| `-test`               | Analyze test files                                                                                                                                                  | `true`  |
| `-generated`          | Analyze generated files                                                                                                                                             | `true`  |
| `-fix`                | Apply suggested fixes (e.g., [naming](#error-naming-conventions) renames, [legacy](#legacy-assertion-query) rewrites)                                               | `false` |
| `-diff`               | With `-fix`, print a unified diff instead of updating the files                                                                                                     | `false` |
| `-overrides <file>`   | Read type overrides from a YAML file (see [Override File](#override-file))                                                                                          | -       |
| `-suggest <file>`     | Append suggestions to an override file (`-` for stdout)                                                                                                             | -       |
| `-heuristics <list>`  | Heuristics to use (`"var"`, `"usage"`, `"receivers"`, `"off"` to disable) _(Experimental)_                                                                          | `all`   |
| `-tracetypes <regex>` | Trace type detection in matching packages _(Experimental)_                                                                                                          | -       |

## Inconsistent Error Type Usage

When an error type `E` is used sometimes as a pointer (`*E`, a _pointer error_) and sometimes directly (`E`, a _value
error_), `errors.AsType` checks for one form miss the other, without any warning from the compiler.

For example, this code ([Go Playground](https://go.dev/play/p/kvaL1A-Pkk9)):

```go
key := []byte("My kung fu is better than yours")
_, err := aes.NewCipher(key)

if keySizeErr, ok := errors.AsType[*aes.KeySizeError](err); ok {
	fmt.Printf("AES keys must be 16, 24, or 32 bytes long, got %d bytes.\n", keySizeErr)
} else if err != nil {
	fmt.Println(err)
}
```

prints a generic error message because [`aes.KeySizeError`](https://pkg.go.dev/crypto/aes#KeySizeError) is a value
error. Running `errortype .` reports:

```console
.../main.go:13:23: Error type "crypto/aes.KeySizeError" should be queried ⏎
   as a value ("errors.AsType[aes.KeySizeError]"), not a pointer. (et:ast)
```

Changing `errors.AsType[*aes.KeySizeError](err)` to `errors.AsType[aes.KeySizeError](err)` fixes this.

A mismatch can also happen on the returning side when `Error` has a value receiver but `Unwrap` (or `Is`, `As`) has a
pointer receiver ([Go Playground](https://go.dev/play/p/g4Lk-pWxkjo)):

```go
type QueryError struct {
	Query string
	Err   error
}

func (e QueryError) Error() string { return fmt.Sprintf("query %q: %v", e.Query, e.Err) }

func (e *QueryError) Unwrap() error { return e.Err }

func userID(name string) (int, error) {
	// ...
		return 0, QueryError{Query: name, Err: ErrNotFound} // not a pointer, has no Unwrap
	// ...
}

func main() {
	_, err := userID("gopher")
	if errors.Is(err, ErrNotFound) { // never matches
		fmt.Println("no such user")
		return
	}

	fmt.Println("unexpected error:", err)
}
```

```console
.../main.go:32:13: Error type "QueryError" should be returned as a pointer ⏎
   ("&QueryError{...}"), not by value. (et:ret+)
```

`Unwrap` has a pointer receiver, so it's only in the [method set](https://go.dev/ref/spec#Method_sets) of `*QueryError`.
`errors.Is` can't unwrap the value error `QueryError{...}` and never finds `ErrNotFound`.

`Error` has a value receiver here, so `errortype` suggests aligning the receivers:

```console
.../main.go:20:6: Method "Error" should be implemented with a pointer receiver ⏎
   ("func (*QueryError) Error"), like "Unwrap" (et:rcv+)
```

If you [declare a value error](#designing-linter-friendly-packages) with `var _ error = QueryError{}`, `errortype`
reports `Unwrap` instead:

```console
.../main.go:22:6: Method "Unwrap" should be implemented with a value receiver ⏎
   ("func (QueryError) Unwrap"), not a pointer receiver (et:rcv)
```

### How Intended Usage Is Detected

The linter determines an error type's intended use (pointer vs. value) by analyzing its _defining_ package. When that
leaves an imported type open, you can settle the form in your own package with a `var _ error = ...` assertion or an
[override file](#override-file). In order of precedence:

1. **Overrides**: User-defined overrides (see [Override File](#override-file)) take the highest priority.

2. **`Unwrap`-related methods**: `Is`, `As`, and `Unwrap` methods with a pointer receiver are only in the method set of
   the pointer. They decide unless a value error is declared (`var _ error = E{}`).

   ```go
   func (e *PointerError) Unwrap() error { /* ... */ } // Suggests pointer type
   ```

3. **Compile-time assertions**: `var _ error = ...` declarations explicitly state intent.

   ```go
   var _ error = ValueError{}         // Asserts ValueError should be used as a value type.
   var _ error = (*PointerError)(nil) // Asserts PointerError should be used as a pointer type.
   ```

4. **Sentinel errors**: Package-level error variables show how the type is used.

   ```go
   var ErrNotFound = &NotFoundError{} // Suggests pointer type
   ```

5. **Usage in functions**: How the type is returned, wrapped with `%w`, or queried determines the form, in that order.
   How it is passed around (call arguments, composite literals, channel sends) counts as weaker evidence.

   ```go
   return ValueError{} // Suggests value type

   return fmt.Errorf("context: %w", ValueError{}) // Suggests value type

   if _, ok := err.(*PointerError); ok { /* ... */ } // Suggests pointer type
   ```

6. **Consistent method receivers**: If all methods have the same receiver type, that style is used.

   > [!NOTE]
   >
   > This heuristic is a fallback and should not be relied upon for defining a type's contract.

### Designing Linter-Friendly Packages

To make intent explicit, add a compile-time assertion in the _defining_ package:

```go
type ValueError struct { /* ... */ }

func (v ValueError) Error() string { /* ... */ }

var _ error = ValueError{} // Explicitly declare value usage.
```

For error types whose `Error() string` method has a pointer receiver, nothing is necessary because the usage is
unambiguous. For types that are not structs, use the zero value, like `var _ error = CodeError(0)` or
`var _ error = MsgError("")`. `errortype` understands any form, including the `var _ error = *new(E)` that `gopls`
offers to add, but the zero value reads better.

When the linter reports ambiguous usage from an imported package that you cannot modify, declare the form in your own
package, or use an [override file](#override-file).

> [!NOTE]
>
> Prefer a pointer receiver on `Error` for pointer error types rather than combining a value receiver with
> `var _ error = (*E)(nil)`; see [Changes in Go 1.27](#changes-in-go-127).

### Pointer vs. Value Errors

Errors travel as [`error`](https://pkg.go.dev/builtin#error) interface values, so indirection is already there. Choosing
between `E` and `*E` is mostly about equality:

- Pointer errors (the usual choice) are equal only if they point to the same instance, like the results of
  [`errors.New`](https://pkg.go.dev/errors#New). [`errors.Is`](https://pkg.go.dev/errors#Is) doesn't match two instances
  even when their fields are the same. Declare `Error` with a pointer receiver, and only `*E` implements `error`.
- Value errors are equal when all their fields match, provided the type is comparable. Use them when identical fields
  should mean identical errors. With a value receiver, both `E` and `*E` implement `error`, so nothing stops a
  constructor from accidentally returning `&E{}`. `errortype` catches this.

## Pointless Comparisons

`errortype` detects comparisons against newly allocated addresses. Per the [Go spec](https://go.dev/ref/spec#Variables),
`&MyStruct{}` and `new(T)` each create a unique address, so `ptr == &MyStruct{}` is `false`. For zero-sized types,
[the result is unspecified](https://go.dev/ref/spec#Comparison_operators) (and not reliably testable).

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

The comparison might still match in two cases, which are only reported with `-strict-is-target`:

- The target's type has an `Is(error) bool` method. An error in the chain might implement custom comparison logic.
- The newly allocated error is the first argument and its type has an `Unwrap() error` method. `errors.Is` traverses its
  chain.

## Error Assertion Queries

`errortype` detects and flags "error assertion queries": functions or methods that check if an error matches a type
using a simple type assertion:

```go
type MyError struct{ msg string }

func (e *MyError) Error() string { return e.msg }

func (e *MyError) Is(err error) bool { // Is-method assertion query
	_, ok := err.(*MyError)
	return ok
}

func IsMyError(err error) bool { // Legacy assertion query
	_, ok := err.(*MyError)
	return ok
}
```

### `Is`-Method Assertion Query

An `Is(err error) bool` method that only asserts the target's type turns `errors.Is` into a type check, which is what
`errors.AsType` is for. Instead of matching a specific instance or logical equivalent, `errors.Is` then matches any
error value of that type:

```go
	err1 := &MyError{"error 1"}
	err2 := &MyError{"error 2"}
	wrapped := fmt.Errorf("wrapped: %w", err1)
	if errors.Is(wrapped, err2) {
		// should not be true
	}
```

This is flagged by default. Remove the `Is` method and use `errors.AsType` when you need to match by type.

> [!NOTE]
>
> `errors.Is` calls a custom `Is(error) bool` method at every step of the error chain, for every target that isn't
> equal. This includes unrelated sentinels in an `if/else` chain. Implement these methods only when necessary.

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
or `errors.As` checks to query the error chain. With `-fix`, the type assertion is rewritten in place:

```go
func IsMyError(err error) bool {
	_, ok := errors.AsType[*MyError](err)
	return ok
}
```

If the helper asserts the wrong form, for example `err.(*NotFoundError)` for a value error type, the rewrite uses the
detected form (`errors.AsType[NotFoundError](err)`), so the fixed helper matches the errors your code returns.

This check is off by default and may be enabled by default in a future release.

## Error Naming Conventions

In Go, names are supposed to communicate intent: the `Err` prefix marks an error sentinel to be matched with
`errors.Is`, and the `Error` suffix marks a structured error type to be queried with `errors.AsType`. Static analyzers
(including `errortype`'s own `-prefix-filter`) can rely on these conventions, too.

When enabled, `errortype` checks that error names follow [Go conventions](https://go.dev/wiki/Errors#naming):

- Sentinel errors (package-level variables) should start with `Err` (or `err` for unexported), e.g., `ErrNotFound`.
- Structured error types should end with `Error`, e.g., `ParseError`.

This check is off by default since it can be noisy on existing codebases. It is part of `-recommended` for development,
and once a codebase is compliant, `-naming` can be enabled in CI to keep it that way.

### Automatic Fixes

The `-fix` flag automatically renames most non-compliant declarations and all their uses in the package, in the spirit
of [`go fix`](https://go.dev/blog/gofix). Generated names are a starting point (for example, `ValidationFailure` becomes
`ValidationFailureError`), so preview the changes with `-fix -diff` and review them before committing.

When an exported declaration is renamed, the old name is kept with a [`// Deprecated:`](https://go.dev/wiki/Deprecated)
comment for backward compatibility, so dependent packages continue to work:

```diff
 var (
-    NotFound = errors.New("not found")
+    // Deprecated: Use [ErrNotFound] instead.
+    NotFound    = ErrNotFound
+    ErrNotFound = errors.New("not found")
 )

-type ErrParse struct{ Line int }
+// Deprecated: Use [ParseError] instead.
+//
+//go:fix inline
+type ErrParse = ParseError
+
+type ParseError struct{ Line int }
```

- Renamed types become aliases marked `//go:fix inline`, as are deprecated `const` sentinels. Dependent packages can be
  [updated automatically](https://go.dev/blog/inliner#source-level-inlining) via `go fix ./...`.
- Renamed `var` sentinels keep the old name as a copy of the new one. `errors.Is` matches either name. Uses in dependent
  packages have to be migrated by hand, and tools like [`staticcheck`](https://staticcheck.dev/docs/checks/#SA1019)
  report them.

## Wrapper Functions

`errortype` automatically detects wrapper functions around [`errors.Is`](https://pkg.go.dev/errors#Is),
[`errors.AsType`](https://pkg.go.dev/errors#AsType), [`errors.As`](https://pkg.go.dev/errors#As), and
[`fmt.Errorf`](https://pkg.go.dev/fmt#Errorf) in your code and its dependencies, and validates their call-site arguments
the same way it does for the standard library functions.

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

## Changes in Go 1.27

In files at `go 1.27` or later, the `printf` check of `go vet` reports a `%w` operand of type `*E` when `E` itself
implements `error` ([#61342](https://go.dev/issue/61342)). This includes code that works, like an error type with a
value receiver that is consistently used as a pointer:

```go
type StatusError struct{ Code int }

func (e StatusError) Error() string { return fmt.Sprintf("status %d", e.Code) }

func get() error {
	return fmt.Errorf("get: %w", &StatusError{404})
}

func main() {
	err := get()
	if err, ok := errors.AsType[*StatusError](err); ok { // matches
		fmt.Println(err)
	}
}
```

To try it, change the `go` directive in the [Go Playground](https://go.dev/play/p/fvUuDfat5G4)'s `go.mod` to `1.27`:

```console
.../main.go:13:26: %w wants operand of error type StatusError, not pointer type *StatusError (defeats errors.Is)
```

The check depends on the Go version of the source file (set by the `go` directive in `go.mod` or a `//go:build go1.27`
constraint), not on the toolchain you run, so the warnings appear once a module raises its `go` directive.

### What `errortype` Reports

`errortype` reports the `Error` receiver, which is where the fix belongs:

```console
.../main.go:10:6: Method "Error" should be implemented with a pointer receiver ⏎
   ("func (*StatusError) Error"), not a value receiver (et:rcv+)
```

With a pointer receiver, only `*StatusError` implements `error`. The `go vet` check no longer applies, and the compiler
rejects converting a non-pointer `StatusError` to `error`.

The receiver is reported:

- with `-receivers` (part of `-recommended`)
- at `go 1.27` or later by default, since `go vet` reports the wraps there

To prepare, run `errortype -recommended ./...` while your module is still at `go 1.26` and fix the receivers it reports.
After raising the `go` directive, `go vet` can still report wraps of types that [only embed an error](#embedded-errors)
and types from dependencies.

### Embedded Errors

`FetchError` embeds `error`, so both `FetchError` and `*FetchError` implement `error`. `go vet` reports wrapping the
pointer:

```go
type FetchError struct {
	error
	URL string
}

func (e *FetchError) Unwrap() error { return e.error }

func fetch(url string) error {
	// ...
	return fmt.Errorf("fetch: %w", &FetchError{error: err, URL: url})
}
```

This works as long as every use is `*FetchError`. `errortype` reports any other use, not the type itself.

Don't wrap `FetchError{...}` instead to satisfy `go vet`. A non-pointer `FetchError` does not have `Unwrap` in its
method set, so `errors.Is` no longer finds the cause, and queries for `*FetchError` don't match.

Consider dropping the embedding, and declare `Error` with a pointer receiver.

### Workarounds

If `Error` is part of an API you can't change, mark it with [`//nolint:errortype`](#suppressing-diagnostics).
`errortype` still checks that every use is `*StatusError`.

To silence `go vet`, convert the operand to `error`:

```go
	return fmt.Errorf("get: %w", error(&StatusError{404}))
```

This only hides the warning, so fix the type when you can. Declare the form with `var _ error = (*StatusError)(nil)`,
because a conversion tells `errortype` less than a `%w` operand.

For types from dependencies, hiding the warning might be the only way. When the dependency leaves the form open,
[declare it in your own package](#override-file).

## Override File

> [!NOTE]
>
> Overrides only apply where you use them. The defining package and other users of the type don't see them, so the same
> error type might end up used as a pointer in one place and as a value in another. When possible, make the intent
> explicit in the defining package (see [Designing Linter-Friendly Packages](#designing-linter-friendly-packages)).

When an imported error type is ambiguous, first declare the form in your own package:

```go
var _ error = (*pkg.StatusError)(nil) // use pkg.StatusError as a pointer in this package
```

This assertion is only used when the defining package leaves the form open, and it applies only to the package that
declares it. Provide an override file when that's not sufficient, for example for unexported types, to ignore a type, or
to force a form that autodetection decided differently.

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

A suggestion makes your code consistent with how the type is _used in your packages_, which may conflict with its
intended design. Refactoring is often preferable to overriding.

### Wrapper Overrides _(Experimental)_

[Wrapper functions](#wrapper-functions) are detected automatically, so this should rarely be needed. For the cases
autodetection misses, the override file accepts a `wrappers` section. The format is experimental and may change between
releases.

```yaml
wrappers:
  is: # like errors.Is
    - your.pkg/errs.Is
  as: # like errors.As
    - (your.pkg/errs.Suite).As
  astype: # like errors.AsType
    - your.pkg/errs.AsType
  errorf: # like fmt.Errorf
    - your.pkg/errs.Errorf
```

Functions are written as `path.Name`, methods as `(path.Type).Name`. The `golangci-lint` plugin accepts the same
`wrappers` setting.

## Suppressing Diagnostics

For deliberate exceptions, like a value receiver on an established API that can't change, the standalone tool honors
`//nolint:errortype`:

- in the doc comment above the `package` clause, for the whole file,
- in the doc comment of a checked declaration: methods, types, legacy helpers, and renamed symbols, or
- at the end of the line for types and variables.

With the [`golangci-lint` plugin](#golangci-lint-module-plugin), `golangci-lint`'s own `//nolint` handling applies.

## Diagnostic Codes

`errortype` ends each diagnostic with a short code, like `(et:wrp)`, that categorizes the issue. The codes are mostly
for internal use and **not stable**; see [CODES.md](CODES.md).

## Integration

### `golangci-lint` Module Plugin

Add `.custom-gcl.yaml` to your project:

```yaml
---
version: v2.15.0

name: golangci-lint
destination: .

plugins:
  - module: fillmore-labs.com/errortype
    import: fillmore-labs.com/errortype/gclplugin
    version: v0.0.12
```

Run `golangci-lint custom` to build a custom executable. Configure in `.golangci.yaml`:

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
        description: checks pointer/value usage, comparisons and naming of Go error types.
        original-url: https://fillmore-labs.com/errortype
        settings:
          # All recommended checks, for local development:
          recommended: true
          # In CI, enable the individual checks your codebase passes instead:
          # naming: true
          # legacy: true
          # receivers: true
          # non-comparable: true
          # style-check: true
          overrides:
            pointer:
              - your.pkg/a.PointerOverrideError
            value:
              - your.pkg/a.ValueOverrideError
            suppress:
              - your.pkg/a.SuppressOverrideError
```

The other [flags](#command-line-flags) are available as settings with the same names (`check-unused`, `prefix-filter`,
`strict-is-method`, `strict-is-target`, `unchecked-assert`). Then run:

```shell
./golangci-lint run .
```

See the [module plugin documentation](https://golangci-lint.run/plugins/module-plugins/#the-automatic-way).

## Related Tools

Since v0.24.0, the Go language server `gopls` ships the
[`ptrtoerror`](https://go.dev/gopls/release/v0.24.0#ptrtoerror-analyzer) analyzer, which is based on the same principle:
an error type should be used either as `E` or as `*E`, never both. It is enabled by default and gives you feedback right
in your editor (e.g., VS Code with the Go extension). It's great to see the problem tackled in the Go tooling itself.

`ptrtoerror` decides from conversions to `error` (including type assertions and `errors.AsType`) in the defining
package. When there are none, it asks you to declare the form with `var _ error = ...`, the same declaration `errortype`
uses. That's a good habit for new error types. The declaration settles the form for `errortype`, too, and for the
packages that import the type.

`errortype` infers the form from how the code uses the type (see
[How Intended Usage Is Detected](#how-intended-usage-is-detected)) and reports what contradicts it. It also:

- sees uses that aren't conversions, like `%w` operands and `errors.As` targets, directly or through
  [wrapper functions](#wrapper-functions);
- reports when the form in use lacks a method, like an `Unwrap` with a pointer receiver on a value error;
- tells you which side of a conflict to fix;
- reports undetermined usage where your code uses an ambiguous dependency type (which `ptrtoerror` reports only in the
  dependency itself), so you can settle the form locally or in an override file;
- runs before you commit and in CI, on code you don't have open in your editor;
- runs in your editor through the [`golangci-lint` module plugin](#golangci-lint-module-plugin).

There's a [proposal](https://go.dev/issue/82742) to publish `ptrtoerror` outside of `gopls`.

## Links

- [Background on the problem this linter solves](https://blog.fillmore-labs.com/posts/errors-1/)
- [Why you shouldn't call `Unwrap` in `Is(error) bool` methods](https://blog.fillmore-labs.com/posts/errors-2/)
- [Go Wiki: Error naming](https://go.dev/wiki/Errors#naming)
- [gopls `ptrtoerror`, detects when a concrete type is converted to the error interface inconsistently](https://pkg.go.dev/golang.org/x/tools/gopls@v0.24.0/internal/analysis/ptrtoerror)

## License

Licensed under the Apache License 2.0. See [LICENSE](LICENSE) for details.
