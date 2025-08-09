package errors_test

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestPanic(t *testing.T) {
	capturePanic := func() (v error) {
		defer func() {
			if r := recover(); r != nil {
				v = r.(error)
			}
		}()
		errors.Panic(fmt.Errorf("just a test: %d", 1337))
		return
	}

	v := capturePanic()

	assert.Contains(t, fmt.Sprintf("%#v", v), "just a test: 1337")
}
