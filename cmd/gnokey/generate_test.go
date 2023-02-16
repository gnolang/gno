package main

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_execGenerateNormal(t *testing.T) {
	t.Parallel()

	cfg := &generateCfg{
		customEntropy: false,
	}

	err := execGenerate(cfg, []string{}, nil)
	require.NoError(t, err)
}

func Test_execGenerateUser(t *testing.T) {
	t.Parallel()
	
	cfg := &generateCfg{
		customEntropy: true,
	}

	err := execGenerate(cfg, []string{}, bufio.NewReader(strings.NewReader("")))
	require.Error(t, err)
	require.Equal(t, err.Error(), "EOF")

	// Try again
	err = execGenerate(cfg, []string{}, bufio.NewReader(strings.NewReader("Hi!\n")))
	require.Error(t, err)
	require.Equal(t, err.Error(),
		"256-bits is 43 characters in Base-64, and 100 in Base-6. You entered 3, and probably want more")

	// Now provide "good" entropy :)
	fakeEntropy := strings.Repeat(":)", 40) + "\ny\n" // entropy + accept count
	err = execGenerate(cfg, []string{}, bufio.NewReader(strings.NewReader(fakeEntropy)))
	require.NoError(t, err)

	// Now provide "good" entropy but no answer
	fakeEntropy = strings.Repeat(":)", 40) + "\n" // entropy + accept count
	err = execGenerate(cfg, []string{}, bufio.NewReader(strings.NewReader(fakeEntropy)))
	require.Error(t, err)

	// Now provide "good" entropy but say no
	fakeEntropy = strings.Repeat(":)", 40) + "\nn\n" // entropy + accept count
	err = execGenerate(cfg, []string{}, bufio.NewReader(strings.NewReader(fakeEntropy)))
	require.NoError(t, err)
}
