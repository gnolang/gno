> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `uassert` - test assertions

Assertion helpers for writing Gno tests, in both `_test.gno` and `_filetest.gno` files. Adapted, lighter port of `stretchr/testify/assert`. Each helper takes a `TestingT`, reports the failure via `t.Errorf`, and lets the test keep running.

## Usage

```go
import (
    "testing"

    "gno.land/p/nt/uassert/v0"
)

func TestAdd(cur realm, t *testing.T) {
    got, err := Add(2, 3)
    uassert.NoError(t, err)
    uassert.Equal(t, 5, got)
    uassert.True(t, got > 0, "result must be positive")

    uassert.PanicsWithMessage(t, cur, "div by zero", func() {
        Div(1, 0)
    })
}
```

Every helper returns a `bool` (`true` on success) so they can be chained or used in conditionals.

## API

```go
type TestingT interface {
    Helper()
    Skip(args ...any)
    Fatalf(fmt string, args ...any)
    Errorf(fmt string, args ...any)
    Logf(fmt string, args ...any)
    Fail()
    FailNow()
}
```

Equality and emptiness (supports `string`, `address`, `bool`, all int/uint widths, `float32/64`):

```go
func Equal(t TestingT, expected, actual any, msgs ...string) bool
func NotEqual(t TestingT, expected, actual any, msgs ...string) bool
func Empty(t TestingT, obj any, msgs ...string) bool
func NotEmpty(t TestingT, obj any, msgs ...string) bool
```

Truthiness and nil:

```go
func True(t TestingT, value bool, msgs ...string) bool
func False(t TestingT, value bool, msgs ...string) bool
func Nil(t TestingT, value any, msgs ...string) bool
func NotNil(t TestingT, value any, msgs ...string) bool
func TypedNil(t TestingT, value any, msgs ...string) bool
func NotTypedNil(t TestingT, value any, msgs ...string) bool
```

Errors:

```go
func NoError(t TestingT, err error, msgs ...string) bool
func Error(t TestingT, err error, msgs ...string) bool
func ErrorContains(t TestingT, err error, contains string, msgs ...string) bool
func ErrorIs(t TestingT, err, target error, msgs ...string) bool
```

Panics and aborts (`f` may be `func()` or `func(realm)`; pass the test's own `cur` as `rlm`):

```go
func PanicsWithMessage(t TestingT, rlm realm, msg string, f any, msgs ...string) bool
func PanicsContains(t TestingT, rlm realm, substr string, f any, msgs ...string) bool
func NotPanics(t TestingT, rlm realm, f any, msgs ...string) bool
func AbortsWithMessage(t TestingT, rlm realm, msg string, f any, msgs ...string) bool
func AbortsContains(t TestingT, rlm realm, substr string, f any, msgs ...string) bool
func NotAborts(t TestingT, rlm realm, f any, msgs ...string) bool
```

## Notes

- A *panic* is a same-realm runtime failure caught with `recover`. An *abort* is a panic that crosses a realm boundary, caught with gno's `revive`. Use the right variant: `PanicsX` for same-realm, `AbortsX` for cross-realm. `NotPanics` covers both.
- `uassert` reports the failure but lets the test continue. Use `gno.land/p/nt/urequire/v0` when subsequent assertions wouldn't be meaningful after a failure.
