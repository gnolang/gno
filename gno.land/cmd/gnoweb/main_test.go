package main

import (
	"flag"
	"testing"
)

func TestFlagHelp(t *testing.T) {
	err := runMain([]string{"-h"})
	if err != flag.ErrHelp {
		t.Errorf("should display usage")
	}
}
