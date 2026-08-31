> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `urequire` - fail-fast test assertions

Sister package to `uassert`. Same assertions, but each one calls `t.FailNow()` on failure so the test stops immediately instead of continuing.

## Usage

```go
import (
    "testing"

    "gno.land/p/nt/urequire/v0"
)

func TestPipeline(t *testing.T) {
    out, err := Build()
    urequire.NoError(t, err)        // aborts the test if Build failed
    urequire.NotNil(t, out)         // out is safe to dereference below
    urequire.Equal(t, "ready", out.Status)
}
```

## API

Helpers take `uassert.TestingT` and return nothing — they either pass or stop the test.

Equality and emptiness:

```go
func Equal(t uassert.TestingT, expected, actual any, msgs ...string)
func NotEqual(t uassert.TestingT, expected, actual any, msgs ...string)
func Empty(t uassert.TestingT, obj any, msgs ...string)
func NotEmpty(t uassert.TestingT, obj any, msgs ...string)
```

Truthiness and nil:

```go
func True(t uassert.TestingT, value bool, msgs ...string)
func False(t uassert.TestingT, value bool, msgs ...string)
func Nil(t uassert.TestingT, value any, msgs ...string)
func NotNil(t uassert.TestingT, value any, msgs ...string)
func TypedNil(t uassert.TestingT, value any, msgs ...string)
func NotTypedNil(t uassert.TestingT, value any, msgs ...string)
```

Errors:

```go
func NoError(t uassert.TestingT, err error, msgs ...string)
func Error(t uassert.TestingT, err error, msgs ...string)
func ErrorContains(t uassert.TestingT, err error, contains string, msgs ...string)
func ErrorIs(t uassert.TestingT, err, target error, msgs ...string)
```

Panics and aborts (`f` may be `func()` or `func(realm)`; pass the test's own `cur` as `rlm`):

```go
func PanicsWithMessage(t uassert.TestingT, rlm realm, msg string, f any, msgs ...string)
func PanicsContains(t uassert.TestingT, rlm realm, substr string, f any, msgs ...string)
func NotPanics(t uassert.TestingT, rlm realm, f any, msgs ...string)
func AbortsWithMessage(t uassert.TestingT, rlm realm, msg string, f any, msgs ...string)
func AbortsContains(t uassert.TestingT, rlm realm, substr string, f any, msgs ...string)
func NotAborts(t uassert.TestingT, rlm realm, f any, msgs ...string)
```

## Notes

- Use `urequire` when the rest of the test depends on the assertion holding (e.g. a `nil` check before dereferencing). Use `uassert` when you want to collect multiple failures from the same test run.
- Each `urequire` helper is a thin wrapper that calls the matching `uassert` helper and then `t.FailNow()` on failure.
