package errors_test

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestPanic(t *testing.T) {
	ctx := errors.FatalContext()
	assert.Nil(t, ctx.Err())

	err := func() error {
		return errors.Fatal(fmt.Errorf("should not happen"))
	}()

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "should not happen")
	assert.NotNil(t, ctx.Err())
	assert.Equal(t, ctx.Err().Error(), "context canceled")
}
