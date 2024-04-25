package main

import (
	"errors"
	"flag"
	"testing"
)

func TestFlagHelp(t *testing.T) {
	err := runMain([]string{"-h"})
	if !errors.Is(err, flag.ErrHelp) {
		t.Errorf("should display usage")
	}
}
