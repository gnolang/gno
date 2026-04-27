package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintDiscoveryBanner_Format(t *testing.T) {
	var buf bytes.Buffer
	printDiscoveryBanner(&buf)

	out := buf.String()
	assert.Contains(t, out, "no workspace")
	assert.Contains(t, out, "gnomod.toml")
	assert.Contains(t, out, "gnowork.toml")
	assert.Contains(t, out, "-extra-root")
	// Banner should be visually distinct (multi-line, not a one-liner).
	assert.GreaterOrEqual(t, strings.Count(out, "\n"), 2)
}
