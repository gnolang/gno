package balances

import (
	"bufio"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/contribs/gnogenesis/internal/common"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getDummyBalances generates dummy balance lines
func getDummyBalances(t *testing.T, count int) []gnoland.Balance {
	t.Helper()

	dummyKeys := common.GetDummyKeys(t, count)
	amount := std.NewCoins(std.NewCoin(ugnot.Denom, 10))

	balances := make([]gnoland.Balance, len(dummyKeys))

	for index, key := range dummyKeys {
		balances[index] = gnoland.Balance{
			Address: key.Address(),
			Amount:  amount,
		}
	}

	return balances
}

func TestGenesis_Balances_Export(t *testing.T) {
	t.Parallel()

	t.Run("invalid genesis file", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := NewBalancesCmd(commands.NewTestIO())
		args := []string{
			"export",
			"--genesis-path",
			"dummy-path",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, common.ErrUnableToLoadGenesis.Error())
	})

	t.Run("invalid genesis app state", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.GetDefaultGenesis()
		genesis.AppState = nil // no app state
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := NewBalancesCmd(commands.NewTestIO())
		args := []string{
			"export",
			"--genesis-path",
			tempGenesis.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, common.ErrAppStateNotSet.Error())
	})

	t.Run("no output file specified", func(t *testing.T) {
		t.Parallel()

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		genesis := common.GetDefaultGenesis()
		genesis.AppState = gnoland.GnoGenesisState{
			Balances: getDummyBalances(t, 1),
		}
		require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

		// Create the command
		cmd := NewBalancesCmd(commands.NewTestIO())
		args := []string{
			"export",
			"--genesis-path",
			tempGenesis.Name(),
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, common.ErrNoOutputFile.Error())
	})

	t.Run("valid balances export", func(t *testing.T) {
		t.Parallel()

		// Generate dummy balances
		balances := getDummyBalances(t, 100)

		tempGenesis, cleanup := testutils.NewTestFile(t)
		t.Cleanup(cleanup)

		// Prepare the output file
		outputFile, outputCleanup := testutils.NewTestFile(t)
		t.Cleanup(outputCleanup)

		saveGenesisFile := func(outputPath string) {
			genesis := common.GetDefaultGenesis()
			genesis.AppState = gnoland.GnoGenesisState{
				Balances: balances,
			}
			require.NoError(t, genesis.SaveAs(tempGenesis.Name()))

			// Create the command
			cmd := NewBalancesCmd(commands.NewTestIO())
			args := []string{
				"export",
				"--genesis-path",
				tempGenesis.Name(),
				outputPath,
			}

			// Run the command
			cmdErr := cmd.ParseAndRun(context.Background(), args)
			require.NoError(t, cmdErr)
		}

		saveGenesisFile(outputFile.Name())
		readIt := func(p string) string {
			blob, err := os.ReadFile(p)
			require.NoError(t, err)
			return string(blob)
		}

		// Validate the transactions were written down
		outputFile.Seek(0, 0) // Seek back to the front of the outputFile.
		scanner := bufio.NewScanner(outputFile)

		outputBalances := make([]gnoland.Balance, 0)
		for scanner.Scan() {
			var balance gnoland.Balance
			err := balance.Parse(scanner.Text())
			require.NoError(t, err)

			outputBalances = append(outputBalances, balance)
		}

		require.NoError(t, scanner.Err())
		assert.Len(t, outputBalances, len(balances))

		// Next ensure that all balances are sorted by address, deterministically.
		for i := 1; i < len(outputBalances); i++ {
			curr := outputBalances[i].Address
			for j := 0; j < i; j++ {
				prev := outputBalances[j].Address
				if addressIsGreater(prev, curr) {
					t.Fatalf("Non-deterministic order of exported balances\n\t[%d](%s)\n>\n\t[%d](%s)", j, prev, i, curr)
				}
			}
		}

		// Lastly compute the checksum and ensure that it is the same each of the N times.
		outputFile.Close()
		firstGenesisPathMD5 := md5SumFromFile(t, outputFile.Name())
		firstGenesis := readIt(outputFile.Name())
		for i := 1; i <= 10; i++ {
			outpath := filepath.Join(t.TempDir(), "out")
			saveGenesisFile(outpath)
			currentMD5 := md5SumFromFile(t, outpath)
			currentGenesis := readIt(outpath)
			require.Equal(t, firstGenesisPathMD5, currentMD5, "Iteration #%d has a different MD5 checksum\n%s\n\n%s", i, firstGenesis, currentGenesis)
		}
	})
}

func md5SumFromFile(t *testing.T, p string) string {
	t.Helper()

	f, err := os.Open(p)
	require.NoError(t, err)
	fi, err := f.Stat()
	require.NoError(t, err)
	require.True(t, fi.Size() > 100, "at least 100 bytes expected")
	defer f.Close()
	h := md5.New()
	io.Copy(h, f)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func addressIsGreater(a, b bft.Address) bool {
	return a.Compare(b) == 1
}
