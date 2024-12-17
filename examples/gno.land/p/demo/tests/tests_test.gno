package tests_test

import (
	"testing"

	"gno.land/p/demo/tests"

	"gno.land/p/demo/uassert"
)

var World = "WORLD"

func TestGetHelloWorld(t *testing.T) {
	// tests.World is 'world'
	s := "hello " + tests.World + World
	const want = "hello worldWORLD"

	uassert.Equal(t, want, s)
}
