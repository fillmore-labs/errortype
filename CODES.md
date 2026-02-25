# Diagnostic Codes

`errortype` ends each diagnostic with a short code, like `(et:wrp)`, that categorizes the issue.

The codes are mostly for internal use:

- We run `errortype` on public repositories and use the codes to see which kinds of diagnostics show up in the wild and
  which are useful.
- The tests match diagnostics by their code, so the wording of a message can change without touching every test.

You can use them too, for example to group the diagnostics of a single run and see which kinds of issues dominate a
large codebase, or to point at a diagnostic in a bug report.

> [!WARNING]
>
> The codes are **not stable**. They are valid only for the exact `errortype` version you run and change whenever that
> helps, without notice. Do not match on them in scripts, CI rules or suppression lists; choose checks with the
> [options](README.md#command-line-flags) instead.

A trailing `+`, like `(et:ret+)`, marks a variant of the same check, for mismatches usually that the pointer form is
expected.

The codes of this version:

| Code     | Name                        | Description                                                                                                                                                                        |
| -------- | --------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `et:arg` | Call Argument Mismatch      | Incorrect error type passed to a function call or conversion.                                                                                                                      |
| `et:asn` | Assignment Mismatch         | Incorrect type in assignment, composite literal, or channel send.                                                                                                                  |
| `et:ast` | Assertion Mismatch          | Incorrect type in a type assertion, type switch, or `errors.As/AsType` call.                                                                                                       |
| `et:cfl` | Contradictory Assertion     | Compile-time assertions (`var _ error = ...`) declare both the pointer and the value form.                                                                                         |
| `et:cmp` | Pointless Error Comparison  | Comparison against `&T{}` in `errors.Is`: always false. Use `errors.As/AsType` instead.                                                                                            |
| `et:eas` | Invalid Argument            | Invalid target to an `errors.As`-like function, **possible panic**.                                                                                                                |
| `et:emb` | Ambiguous Usage             | Could not determine if error is pointer or value type; use an [override](README.md#override-file).                                                                                 |
| `et:equ` | Pointless Comparison        | Pointer compared against `&T{}`: always false (unspecified for zero-sized types).                                                                                                  |
| `et:err` | Argument Mismatch           | Incorrect target passed to an `errors.As`-like function.                                                                                                                           |
| `et:ias` | `Is`-Method Assertion Query | `Is` method implements an error assertion query.                                                                                                                                   |
| `et:ins` | Instantiation Mismatch      | Error type used in the wrong form as a type argument of a generic function or type.                                                                                                |
| `et:lgc` | Legacy Assertion Query      | Legacy error assertion query does not account for wrapped errors.                                                                                                                  |
| `et:nam` | Error Naming                | Error variables should start with `Err` or `err`; structured error type names should end with `Error`.                                                                             |
| `et:nce` | Non-comparable Value Type   | A [non-comparable](https://go.dev/ref/spec#Comparison_operators) error should be a pointer type.                                                                                   |
| `et:rcv` | Receiver Mismatch           | Method receiver does not match the error type's form: value `Error` with pointer `Unwrap`; with `-receivers`, also `Error` of pointer errors and exported methods of value errors. |
| `et:ret` | Return Mismatch             | Error type returned incorrectly (value as pointer or vice versa).                                                                                                                  |
| `et:sig` | Wrong Signature             | `Unwrap`-related method has a wrong signature (also flagged by [`stdmethods`](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/stdmethods)).                               |
| `et:sty` | Style Mismatch              | Target to `errors.As` is not an address operation; declare a variable for clarity.                                                                                                 |
| `et:uca` | Unchecked Type Assert       | Unchecked type assert might panic on wrapped error; prefer `errors.As/AsType`.                                                                                                     |
| `et:unu` | Unused Result               | Result of `errors.Is/As/AsType`, `fmt.Errorf` (or wrapper) is unused.                                                                                                              |
| `et:unw` | Calling Unwrap              | Unwrapping function called inside `Is(error) bool`; use shallow comparison (direct type assertion) instead.                                                                        |
| `et:var` | Variable Mismatch           | Incorrect type in a variable declaration; only variables named like errors are checked when `-prefix-filter` is on (default).                                                      |
| `et:wrp` | Wrap Mismatch               | Error wrapped as a pointer instead of a value (or vice versa) by `%w` in `fmt.Errorf`.                                                                                             |
