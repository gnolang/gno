package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolvePath(t *testing.T) {
	t.Parallel()

	if os.PathSeparator != '/' {
		t.Skip("ResolvePath test is only written of UNIX-like filesystems")
	}
	wd, err := os.Getwd()
	require.NoError(t, err)
	tt := []struct {
		output  string
		dstPath string
		result  string
	}{
		{
			"transpile-result",
			"./examples/test/test1.gno.gen.go",
			"transpile-result/examples/test/test1.gno.gen.go",
		},
		{
			"/transpile-result",
			"./examples/test/test1.gno.gen.go",
			"/transpile-result/examples/test/test1.gno.gen.go",
		},
		{
			"/transpile-result",
			"/home/gno/examples/test/test1.gno.gen.go",
			"/transpile-result/home/gno/examples/test/test1.gno.gen.go",
		},
		{
			"result",
			"../hello",
			filepath.Join("result", filepath.Join(wd, "../hello")),
		},
	}

	for _, tc := range tt {
		res, err := ResolvePath(tc.output, tc.dstPath)
		// ResolvePath should error only in case we can't get the abs path;
		// so never in normal conditions.
		require.NoError(t, err)
		assert.Equal(t,
			tc.result, res,
			"unexpected result of ResolvePath(%q, %q)", tc.output, tc.dstPath,
		)
	}
}
