package math

import (
	"testing"

	"gno.land/p/demo/uassert"
)

func TestAdd(t *testing.T) {
	uassert.Equal(t, Add(1, 2), 3)
}

func TestSubtract(t *testing.T) {
	uassert.Equal(t, Subtract(5, 3), 2)
}

func TestMultiply(t *testing.T) {
	uassert.Equal(t, Multiply(2, 3), 6)
}
