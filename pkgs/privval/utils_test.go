package privval

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gnolang/gno/pkgs/errors"
)

func TestIsConnTimeoutForNonTimeoutErrors(t *testing.T) {
	assert.False(t, IsConnTimeout(errors.Wrap(ErrDialRetryMax, "max retries exceeded")))
	assert.False(t, IsConnTimeout(errors.New("completely irrelevant error")))
}
