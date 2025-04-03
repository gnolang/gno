package client

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/require"
)

func Test_execGenerateNormal(t *testing.T) {
	t.Parallel()

	cfg := &GenerateCfg{
		RootCfg:       &BaseCfg{},
		CustomEntropy: false,
	}

	err := execGenerate(cfg, []string{}, commands.NewTestIO())
	require.NoError(t, err)
}

func Test_execGenerateUser(t *testing.T) {
	t.Parallel()

	cfg := &GenerateCfg{
		RootCfg:       &BaseCfg{},
		CustomEntropy: true,
	}

	io := commands.NewTestIO()
	io.SetIn(strings.NewReader(""))

	err := execGenerate(cfg, []string{}, io)
	require.Error(t, err)
	require.Equal(t, err.Error(), "EOF")

	// Try again
	io.SetIn(strings.NewReader("Hi!\n"))
	err = execGenerate(cfg, []string{}, io)
	require.Error(t, err)
	require.Equal(t, err.Error(),
		"256-bits is 43 characters in Base-64, and 100 in Base-6. You entered 3, and probably want more")

	// Now provide "good" entropy :)
	fakeEntropy := strings.Repeat(":)", 40) + "\ny\n" // entropy + accept count
	io.SetIn(strings.NewReader(fakeEntropy))
	err = execGenerate(cfg, []string{}, io)
	require.NoError(t, err)

	// Now provide "good" entropy but no answer
	fakeEntropy = strings.Repeat(":)", 40) + "\n" // entropy + accept count
	io.SetIn(strings.NewReader(fakeEntropy))
	err = execGenerate(cfg, []string{}, io)
	require.Error(t, err)

	// Now provide "good" entropy but say no
	fakeEntropy = strings.Repeat(":)", 40) + "\nn\n" // entropy + accept count
	io.SetIn(strings.NewReader(fakeEntropy))
	err = execGenerate(cfg, []string{}, io)
	require.NoError(t, err)
}
