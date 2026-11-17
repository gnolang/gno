package integration

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/amino"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

func TestSimulateSendsBlankSignatures(t *testing.T) {
	const chainID = "tendermint_test"

	cfg := TestingMinimalNodeConfig(gnoenv.RootDir())
	n, remote := TestingInMemoryNode(t, log.NewNoopLogger(), cfg)
	defer n.Stop()
	nodeURL := strings.Replace(remote, "tcp://", "http://", 1)

	home := t.TempDir()
	kb, err := keys.NewKeyBaseFromDir(home)
	require.NoError(t, err)
	pk, err := GeneratePrivKeyFromMnemonic(DefaultAccount_Seed, "", 0, 0)
	require.NoError(t, err)
	require.NoError(t, kb.ImportPrivKey(DefaultAccount_Name, pk, ""))
	signer := crypto.MustAddressFromString(DefaultAccount_Address)

	cli, err := rpcclient.NewHTTPClient(remote)
	require.NoError(t, err)

	makeTx := func(t *testing.T, remote, simulate string, msgs ...std.Msg) *ctypes.ResultBroadcastTxCommit {
		t.Helper()
		const gasWanted = int64(10_000_000)
		fee := std.NewCoin("ugnot", gasWanted)
		mcfg := &client.MakeTxCfg{
			RootCfg:   &client.BaseCfg{BaseOptions: client.BaseOptions{Home: home, Remote: remote}},
			GasWanted: gasWanted,
			GasFee:    fee.String(),
			Broadcast: true,
			Simulate:  simulate,
			ChainID:   chainID,
		}
		tx := std.Tx{Msgs: msgs, Fee: std.NewFee(gasWanted, fee)}
		res, err := client.SignAndBroadcastHandler(mcfg, DefaultAccount_Name, tx, "", commands.NewTestIO())
		require.NoError(t, err)
		return res
	}

	replay := func(t *testing.T, tx std.Tx) *ctypes.ResultBroadcastTxCommit {
		t.Helper()
		res, err := cli.BroadcastTxCommit(context.Background(), amino.MustMarshal(tx))
		require.NoError(t, err)
		return res
	}

	t.Run("maketx simulate only", func(t *testing.T) {
		p := newSimCaptureProxy(t, nodeURL)
		to := crypto.AddressFromPreimage([]byte("dry-run recipient"))
		makeTx(t, p.srv.URL, client.SimulateOnly,
			bank.NewMsgSend(signer, to, std.MustParseCoins("4242ugnot")))

		seen := p.onlySimulated(t)
		require.NotNil(t, seen.Signatures[0].PubKey)
		require.NotEmpty(t, seen.Signatures[0].Signature)
		require.Equal(t, make([]byte, len(seen.Signatures[0].Signature)), seen.Signatures[0].Signature)
		require.Zero(t, p.broadcastCount())
		seen.Fee.GasWanted = 10_000_000 // what was signed; simulation raised it to MaxGas
		require.Error(t, replay(t, seen).CheckTx.Error)
	})

	t.Run("broadcast dry run with stored multisig key", func(t *testing.T) {
		k1, k2 := secp256k1.GenPrivKey(), secp256k1.GenPrivKey()
		pubs := []crypto.PubKey{k1.PubKey(), k2.PubKey()}
		mpk := multisig.NewPubKeyMultisigThreshold(2, pubs)
		maddr := mpk.Address()

		res := makeTx(t, remote, client.SimulateSkip,
			bank.NewMsgSend(signer, maddr, std.MustParseCoins("100000000ugnot")))
		require.NoError(t, res.DeliverTx.Error)

		sign := func(t *testing.T, tx *std.Tx, includePubKey bool) {
			t.Helper()
			acc := queryBaseAccount(t, cli, maddr)
			signBytes, err := tx.GetSignBytes(chainID, acc.AccountNumber, acc.Sequence)
			require.NoError(t, err)
			ms := multisig.NewMultisig(len(pubs))
			for _, key := range []crypto.PrivKey{k1, k2} {
				sig, err := key.Sign(signBytes)
				require.NoError(t, err)
				require.NoError(t, ms.AddSignatureFromPubKey(sig, key.PubKey(), pubs))
			}
			var pub crypto.PubKey
			if includePubKey {
				pub = mpk
			}
			tx.Signatures = []std.Signature{{PubKey: pub, Signature: ms.Marshal()}}
		}

		first := std.Tx{
			Msgs: []std.Msg{bank.NewMsgSend(maddr, maddr, std.MustParseCoins("1ugnot"))},
			Fee:  std.NewFee(10_000_000, std.NewCoin("ugnot", 10_000_000)),
		}
		sign(t, &first, true)
		res = replay(t, first)
		require.NoError(t, res.DeliverTx.Error)

		tx := std.Tx{
			Msgs: []std.Msg{bank.NewMsgSend(maddr, maddr, std.MustParseCoins("1ugnot"))},
			Fee:  std.NewFee(10_000_000, std.NewCoin("ugnot", 10_000_000)),
		}
		sign(t, &tx, false)
		file := filepath.Join(t.TempDir(), "tx.json")
		require.NoError(t, os.WriteFile(file, amino.MustMarshalJSON(tx), 0o600))

		p := newSimCaptureProxy(t, nodeURL)
		cmd := client.NewBroadcastCmd(
			&client.BaseCfg{BaseOptions: client.BaseOptions{Remote: p.srv.URL}},
			commands.NewTestIO(),
		)
		require.NoError(t, cmd.ParseAndRun(context.Background(), []string{"-dry-run", file}))

		seen := p.onlySimulated(t)
		require.Nil(t, seen.Signatures[0].PubKey)
		require.Len(t, amino.MustMarshal(seen), len(amino.MustMarshal(tx)))
		var ms multisig.Multisignature
		require.NoError(t, amino.Unmarshal(seen.Signatures[0].Signature, &ms))
		require.Len(t, ms.Sigs, 2)
		for _, sig := range ms.Sigs {
			require.Equal(t, make([]byte, len(sig)), sig)
		}
		require.Zero(t, p.broadcastCount())
		require.Error(t, replay(t, seen).CheckTx.Error)
	})
}

type simCaptureProxy struct {
	srv *httptest.Server
	mu  sync.Mutex

	simulated  [][]byte
	broadcasts int
}

func (p *simCaptureProxy) onlySimulated(t *testing.T) std.Tx {
	t.Helper()
	p.mu.Lock()
	defer p.mu.Unlock()
	require.Len(t, p.simulated, 1)
	var tx std.Tx
	require.NoError(t, amino.Unmarshal(p.simulated[0], &tx))
	require.Len(t, tx.Signatures, 1)
	return tx
}

func (p *simCaptureProxy) broadcastCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.broadcasts
}

func newSimCaptureProxy(t *testing.T, nodeURL string) *simCaptureProxy {
	t.Helper()
	target, err := url.Parse(nodeURL)
	require.NoError(t, err)
	proxy := httputil.NewSingleHostReverseProxy(target)
	p := &simCaptureProxy{}
	p.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		r.Body = io.NopCloser(bytes.NewReader(body))

		var req rpctypes.RPCRequest
		require.NoError(t, json.Unmarshal(body, &req))
		switch {
		case strings.HasPrefix(req.Method, "broadcast_tx"):
			p.mu.Lock()
			p.broadcasts++
			p.mu.Unlock()
		case req.Method == "abci_query":
			var params struct {
				Path string `json:"path"`
				Data string `json:"data"`
			}
			require.NoError(t, json.Unmarshal(req.Params, &params))
			if params.Path == ".app/simulate" {
				bz, err := base64.StdEncoding.DecodeString(params.Data)
				require.NoError(t, err)
				p.mu.Lock()
				p.simulated = append(p.simulated, bz)
				p.mu.Unlock()
			}
		}
		proxy.ServeHTTP(w, r)
	}))
	t.Cleanup(p.srv.Close)
	return p
}

func queryBaseAccount(t *testing.T, cli *rpcclient.RPCClient, addr crypto.Address) std.BaseAccount {
	t.Helper()
	res, err := cli.ABCIQuery(context.Background(), "auth/accounts/"+addr.String(), nil)
	require.NoError(t, err)
	var acc gnoland.GnoAccount
	require.NoError(t, amino.UnmarshalJSON(res.Response.Data, &acc))
	return acc.BaseAccount
}
