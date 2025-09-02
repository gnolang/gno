package gnolang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_asPrimitive(t *testing.T) {
	for i := PrimitiveType(1); i < 1<<30; i <<= 1 {
		assert.Equal(t, i, asPrimitive(i))
	}

	assert.Equal(t, InvalidType, asPrimitive(nil))
	assert.Equal(t, InvalidType, asPrimitive(&DeclaredType{}))
}
