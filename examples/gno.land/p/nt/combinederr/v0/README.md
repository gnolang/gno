> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `combinederr` - Aggregate multiple errors

Collect several errors into a single `error` value with a semicolon-separated message. Useful for batch operations where you want to report every failure instead of bailing on the first.

## Usage

```go
import "gno.land/p/nt/combinederr/v0"

ce := &combinederr.CombinedError{}
ce.Add(validateName(name))   // nil is silently skipped
ce.Add(validateEmail(email))
ce.Add(validateAge(age))

if ce.Size() > 0 {
    return ce // "invalid name; bad email; age must be > 0"
}
return nil
```

## API

```go
// CombinedError aggregates multiple errors into a single error value.
type CombinedError struct { /* ... */ }

// Add appends err to the combined error. Nil errors are ignored.
func (e *CombinedError) Add(err error)

// Error returns all collected errors joined with "; ".
func (e *CombinedError) Error() string

// Size returns the number of collected errors.
func (e *CombinedError) Size() int
```

## Notes

- A zero-value `CombinedError{}` with no errors added has `Size() == 0` and `Error() == ""`.
- The pointer receiver on `Add` means you must use `&CombinedError{}`.
- Aggregation is message-only: there is no `Unwrap`, so `errors.Is`/`errors.As` never see the collected errors. Use it when you want one human-readable combined string, not typed error matching.
