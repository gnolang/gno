package std

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ZAccount(t *testing.T) {
	zA := ZeroAddress()
	assert.Exactly(t, zA.String(), "g100000000000000000000000000000000dnmcnx")
}
