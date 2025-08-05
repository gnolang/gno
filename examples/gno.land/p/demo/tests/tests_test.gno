package tests_test

import (
	"testing"

	"gno.land/p/demo/tests"
)

var World = "WORLD"

func TestGetHelloWorld(t *testing.T) {
	// tests.World is 'world'
	s := "hello " + tests.World + World
	const want = "hello worldWORLD"
	if s != want {
		t.Error("not the same")
	}
}
