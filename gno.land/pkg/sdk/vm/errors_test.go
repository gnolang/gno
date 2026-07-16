package vm

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	tmerrors "github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"go.uber.org/multierr"
)

// The Error component of ABCIResult is merkle-hashed into the block's
// LastResultsHash, and go/types diagnostic text varies across Go toolchains.
// So the hashed bytes of a rejected type-check must not depend on the
// diagnostic text: any two type-check failures must encode identically, with
// the full messages preserved only on the unhashed path (Result.Log, which
// formats the error with %#v).
func TestErrTypeCheckCoarseHashedError(t *testing.T) {
	errA := ErrTypeCheck(errors.New("a.gno:1:1: undefined: foo"))
	errB := ErrTypeCheck(multierr.Combine(
		errors.New("b.gno:9:9: cannot range over p: requires go1.23 or later"),
		errors.New("b.gno:12:1: declared and not used: x"),
	))

	resA := bft.ABCIResult{Error: sdk.ABCIError(errA)}
	resB := bft.ABCIResult{Error: sdk.ABCIError(errB)}
	assert.Equal(t, resA.Bytes(), resB.Bytes(),
		"hashed result bytes must be independent of type-check diagnostic text")

	// The hashed component is the bare sentinel.
	assert.Equal(t, TypeCheckError{}, tmerrors.Cause(errA))

	// The full diagnostics survive on the unhashed path.
	assert.Contains(t, fmt.Sprintf("%#v", errA), "undefined: foo")
	assert.Contains(t, fmt.Sprintf("%#v", errB), "requires go1.23")
	assert.Contains(t, fmt.Sprintf("%#v", errB), "declared and not used: x")
}
