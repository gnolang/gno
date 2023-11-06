package tests_test

import (
	"testing"

	"gno.land/p/demo/tests"
)

func TestGetHelloWorld(t *testing.T) {
	s := "hello " + tests.World
	println(s)
}
