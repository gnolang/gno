package main

import (
	"strings"
	"testing"
)

func Test_execDoctest_InvalidPath(t *testing.T) {
	cfg := &doctestCfg{
		path:  "",
		index: 0,
	}
	args := []string{}

	err := execDoctest(cfg, args)
	if err == nil || !strings.Contains(err.Error(), "missing markdown-path flag") {
		t.Errorf("execDoctest should fail with missing path error, got: %v", err)
	}
}
