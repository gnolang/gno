package gnoclient

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
)

const (
	govdaoTakeover = `package main

import "gno.land/r/gov/dao"

func main(cur realm) {
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(nil, []string{"gno.land/r/evil/dao"}))
	println("allowlist now:", dao.AllowedDAOs()[0])
}`

	govdaoLockdown = `package main

import "gno.land/r/gov/dao"

func main(cur realm) {
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(nil, []string{"gno.land/r/gov/dao/impl/v0"}))
}`
)

// TestGovDAOAllowlistAfterGenesis runs a takeover MsgRun after genesis against
// a real node. An unset allowlist admits every caller only at height 0 and on
// the dev chain IDs (dev, tendermint_test); the txtar harness always runs as
// tendermint_test, so only a node with another chain ID reaches the closed
// branch.
func TestGovDAOAllowlistAfterGenesis(t *testing.T) {
	for _, tc := range []struct {
		name            string
		chainID         string
		genesisLockdown bool
		wantErr         string // empty: the takeover succeeds
	}{
		{"staging/no-lockdown", "staging", false, "permission denied"},
		{"staging/genesis-lockdown", "staging", true, "permission denied"},
		{"tendermint_test/no-lockdown", "tendermint_test", false, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res, err := runGovDAOTakeover(t, tc.chainID, tc.genesisLockdown)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Contains(t, string(res.DeliverTx.Data), "allowlist now: gno.land/r/evil/dao")
		})
	}
}

func runGovDAOTakeover(t *testing.T, chainID string, genesisLockdown bool) (*ctypes.ResultBroadcastTxCommit, error) {
	t.Helper()

	rootdir := gnoenv.RootDir()
	config := integration.TestingMinimalNodeConfig(rootdir)
	config.SkipGenesisSigVerification = true
	config.Genesis.ChainID = chainID
	state := config.Genesis.AppState.(gnoland.GnoGenesisState)
	state.Txs = append(state.Txs, loadpkgs(t, rootdir, "gno.land/r/gov/dao/loader/v0")...)
	if genesisLockdown {
		privKey, err := integration.GeneratePrivKeyFromMnemonic(integration.DefaultAccount_Seed, "", 0, 0)
		require.NoError(t, err)
		lockdown := []gnoland.TxWithMetadata{{Tx: std.Tx{
			Msgs: []std.Msg{vm.MsgRun{
				Caller:  privKey.PubKey().Address(),
				Package: &std.MemPackage{Name: "main", Files: []*std.MemFile{{Name: "main.gno", Body: govdaoLockdown}}},
			}},
			Fee: std.NewFee(50000000, std.MustParseCoin(ugnot.ValueString(1000000))),
		}}}
		require.NoError(t, gnoland.SignGenesisTxs(lockdown, privKey, chainID))
		state.Txs = append(state.Txs, lockdown...)
	}
	config.Genesis.AppState = state
	// A genesis tx that fails, such as a lockdown the height-0 window refused,
	// stops the node instead of being skipped.
	config.InitChainerConfig.GenesisTxResultHandler = gnoland.PanicOnFailingTxResultHandler

	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)
	client := Client{Signer: newInMemorySigner(t, chainID), RPCClient: rpcClient}
	caller, err := client.Signer.Info()
	require.NoError(t, err)
	acc, _, err := client.QueryAccount(caller.GetAddress())
	require.NoError(t, err)

	return client.Run(BaseTxCfg{
		GasFee:         ugnot.ValueString(2100000),
		GasWanted:      50000000,
		AccountNumber:  acc.AccountNumber,
		SequenceNumber: acc.Sequence,
	}, vm.MsgRun{
		Caller:  caller.GetAddress(),
		Package: &std.MemPackage{Name: "main", Files: []*std.MemFile{{Name: "main.gno", Body: govdaoTakeover}}},
	})
}
