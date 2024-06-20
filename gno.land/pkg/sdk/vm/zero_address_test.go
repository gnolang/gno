package vm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ZAccount(t *testing.T) {
	zA := ZeroAddress()
	assert.Exactly(t, zA.String(), "g1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqluuxe")
}
