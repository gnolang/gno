> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `testutils` - misc testing helpers

Small grab-bag of helpers for `_test.gno` files: deterministic fake addresses, a call-stack wrapper, and fixtures exercising access rules (exported/unexported fields, methods, and interfaces).

## Usage

```go
import (
    "testing"

    "gno.land/p/nt/testutils/v0"
)

func TestTransfer(t *testing.T) {
    alice := testutils.TestAddress("alice") // deterministic g1... address
    bob := testutils.TestAddress("bob")

    testutils.WrapCall(func() {
        Transfer(alice, bob, 100)
    })
}
```

## API

Addresses:

```go
// TestAddress returns a deterministic bech32 g1... address derived from name.
// name must be at most 20 bytes; it is right-padded with '_' before encoding.
func TestAddress(name string) address
```

Call-stack helper:

```go
// WrapCall invokes fn after adding one extra frame to the call stack.
// Useful for tests that inspect caller depth.
func WrapCall(fn func())
```

Access-rule fixtures (used by VM file tests to exercise exported vs. unexported visibility):

```go
type TestAccessStruct struct {
    PublicField  string
    // privateField is unexported on purpose.
}

func NewTestAccessStruct(pub, priv string) TestAccessStruct
func (TestAccessStruct) PublicMethod() string

type PrivateInterface interface {
    // unexported method — only satisfiable from within this package.
}

func PrintPrivateInterface(pi PrivateInterface)

var TestVar1 int // initialized to 123 in init()
```

## Notes

- `TestAddress` panics if `name` exceeds 20 bytes; the resulting address is reproducible across runs, which is what you want in tests.
- The access fixtures exist mainly to back GnoVM file tests under `gnovm/tests/files/`; most user code only needs `TestAddress`.
