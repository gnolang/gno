package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gnolang/faucet/spec"
	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Faucet settings for the end-to-end test, in the units gnofaucet's flags take,
// and the amount its drips request.
const (
	e2eMaxSendAmount = "300000000ugnot"
	e2eGasFee        = "1000000ugnot"
	e2eGasWanted     = "5000000"
	e2eDripAmount    = "100000000ugnot"
)

// Waits of the end-to-end test.
const (
	e2eFaucetStartTimeout = 10 * time.Second
	e2eFaucetStopTimeout  = 10 * time.Second
	e2ePollInterval       = 5 * time.Millisecond
	e2eHealthCheckTimeout = time.Second
	e2eRequestTimeout     = 30 * time.Second
)

// TestCaptchaFaucetAgainstARealChain runs the faucet of `gnofaucet serve
// captcha` against an in-memory gno.land node built from this repository, sends
// JSON-RPC drips to it over HTTP, with the captcha token in meta as faucet-hub
// sends it, and checks the outcome on-chain.
//
// The unit tests exercise each layer against fakes. What they cannot show is
// that a drip works against a real node: that the faucet decodes the accounts
// and transaction results the node returns, and that the node accepts the
// transactions the faucet signs. Each of those can break while every unit test
// passes.
//
// hCaptcha is verified against its real API with its documented test keys, so
// the test needs network access and is skipped in short mode.
func TestCaptchaFaucetAgainstARealChain(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end-to-end test in short mode: hCaptcha verification needs network access")
	}

	chain := startTestChain(t)

	// The subtests run sequentially: they share the faucet account, and the
	// faucet does not serialize transactions sent from one account.

	t.Run("drips the requested amount", func(t *testing.T) {
		faucetURL := startTestFaucet(t, chain)
		to := newTestAddress()

		status, res := sendDrip(t, faucetURL, []string{to.String(), e2eDripAmount}, hcaptchaTestResponse)

		require.Equal(t, http.StatusOK, status)
		require.Nil(t, res.Error)
		assert.Equal(t, "successfully executed faucet transfer", res.Result)
		assert.Equal(t, std.MustParseCoins(e2eDripAmount), chain.balance(t, to))
	})

	t.Run("drips the maximum amount when none is requested", func(t *testing.T) {
		faucetURL := startTestFaucet(t, chain)
		to := newTestAddress()

		status, res := sendDrip(t, faucetURL, []string{to.String()}, hcaptchaTestResponse)

		require.Equal(t, http.StatusOK, status)
		require.Nil(t, res.Error)
		assert.Equal(t, std.MustParseCoins(e2eMaxSendAmount), chain.balance(t, to))
	})

	t.Run("rejects an amount above the maximum", func(t *testing.T) {
		faucetURL := startTestFaucet(t, chain)
		to := newTestAddress()
		aboveMax := std.MustParseCoins(e2eMaxSendAmount).Add(std.NewCoins(std.NewCoin("ugnot", 1)))

		status, res := sendDrip(t, faucetURL, []string{to.String(), aboveMax.String()}, hcaptchaTestResponse)

		require.Equal(t, http.StatusOK, status)
		require.NotNil(t, res.Error)
		assert.Equal(t, spec.InvalidRequestErrorCode, res.Error.Code)
		assert.True(t, chain.balance(t, to).IsZero())
	})

	t.Run("rejects an invalid captcha", func(t *testing.T) {
		faucetURL := startTestFaucet(t, chain)
		to := newTestAddress()

		status, res := sendDrip(t, faucetURL, []string{to.String(), e2eDripAmount}, "not-a-valid-hcaptcha-token")

		require.Equal(t, http.StatusOK, status)
		require.NotNil(t, res.Error)
		assert.Equal(t, spec.InvalidParamsErrorCode, res.Error.Code)
		assert.True(t, chain.balance(t, to).IsZero())
	})

	t.Run("throttles a second request from the same IP", func(t *testing.T) {
		faucetURL := startTestFaucet(t, chain)
		to := newTestAddress()

		status, res := sendDrip(t, faucetURL, []string{to.String(), e2eDripAmount}, hcaptchaTestResponse)
		require.Equal(t, http.StatusOK, status)
		require.Nil(t, res.Error)

		status, _ = sendDrip(t, faucetURL, []string{to.String(), e2eDripAmount}, hcaptchaTestResponse)

		assert.Equal(t, http.StatusUnauthorized, status)
		assert.Equal(t, std.MustParseCoins(e2eDripAmount), chain.balance(t, to))
	})
}

// testChain is an in-memory gno.land node that the faucet under test drips
// from. Its genesis funds integration.DefaultAccount_Address.
type testChain struct {
	remote  string // RPC address in the http:// form gnofaucet's -remote takes
	chainID string
	queries gnoclient.Client
}

// startTestChain starts an in-memory gno.land node for the test and stops it
// when the test ends.
func startTestChain(t *testing.T) testChain {
	t.Helper()

	cfg := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
	node, remote := integration.TestingInMemoryNode(t, log.NewNoopLogger(), cfg)
	t.Cleanup(func() {
		assert.NoError(t, node.Stop())
	})

	rpc, err := rpcclient.NewHTTPClient(remote)
	require.NoError(t, err)

	return testChain{
		remote:  strings.Replace(remote, "tcp://", "http://", 1),
		chainID: cfg.Genesis.ChainID,
		queries: gnoclient.Client{RPCClient: rpc},
	}
}

// balance returns every coin address holds on the chain.
func (c testChain) balance(t *testing.T, address crypto.Address) std.Coins {
	t.Helper()

	coins, _, err := c.queries.QueryBalance(address)
	require.NoError(t, err)

	return coins
}

// startTestFaucet runs the captcha faucet against chain, configured through its
// command-line flags with the faucet account funded by the chain's genesis, and
// returns the faucet URL once it answers /health. The faucet stops when the
// test ends. Its logs are captured, and an error logged by the faucet fails the
// test.
func startTestFaucet(t *testing.T, chain testChain) string {
	t.Helper()

	listenAddress := freeListenAddress(t)

	cfg := &captchaCfg{rootCfg: &serveCfg{}}
	fs := flag.NewFlagSet("captcha", flag.ContinueOnError)
	cfg.rootCfg.RegisterFlags(fs)
	cfg.RegisterFlags(fs)
	require.NoError(t, fs.Parse([]string{
		"-listen-address", listenAddress,
		"-remote", chain.remote,
		"-chain-id", chain.chainID,
		"-mnemonic", integration.DefaultAccount_Seed,
		"-max-send-amount", e2eMaxSendAmount,
		"-gas-fee", e2eGasFee,
		"-gas-wanted", e2eGasWanted,
		"-captcha-secret", hcaptchaTestSecret,
	}))

	logs := &syncBuffer{}
	testIO := commands.NewTestIO()
	testIO.SetOut(commands.WriteNopCloser(logs))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	var serveErr error
	go func() {
		defer close(done)
		serveErr = execCaptcha(ctx, cfg, testIO)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
			assert.NoError(t, serveErr)
			assertNoErrorLogged(t, logs.String())
		case <-time.After(e2eFaucetStopTimeout):
			t.Errorf("faucet did not stop within %s", e2eFaucetStopTimeout)
		}
	})

	faucetURL := "http://" + listenAddress
	httpClient := http.Client{Timeout: e2eHealthCheckTimeout}
	require.Eventually(t, func() bool {
		select {
		case <-done:
			// The faucet exited; the check below reports why.
			return true
		default:
		}

		res, err := httpClient.Get(faucetURL + "/health")
		if err != nil {
			return false
		}
		defer res.Body.Close()

		return res.StatusCode == http.StatusOK
	}, e2eFaucetStartTimeout, e2ePollInterval)

	select {
	case <-done:
		require.FailNow(t, "faucet exited before serving", "error: %v", serveErr)
	default:
	}

	return faucetURL
}

// assertNoErrorLogged checks that logs, the faucet's JSON log output, holds at
// least one entry and no entry at error level.
func assertNoErrorLogged(t *testing.T, logs string) {
	t.Helper()

	logs = strings.TrimSpace(logs)
	require.NotEmpty(t, logs, "the faucet logged nothing")

	for line := range strings.SplitSeq(logs, "\n") {
		var entry struct {
			Level string `json:"level"`
		}
		if !assert.NoError(t, json.Unmarshal([]byte(line), &entry), "log line %q is not a JSON entry", line) {
			continue
		}

		assert.NotEqual(t, "error", entry.Level, "the faucet logged an error: %s", line)
	}
}

// syncBuffer is a bytes.Buffer that the faucet's logger can write to from
// concurrent request handlers.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}

// freeListenAddress returns a loopback address with a port that is free when
// the function returns.
func freeListenAddress(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	address := listener.Addr().String()
	require.NoError(t, listener.Close())

	return address
}

// newTestAddress returns the address of a fresh key, which holds nothing on any
// chain.
func newTestAddress() crypto.Address {
	return secp256k1.GenPrivKey().PubKey().Address()
}

// sendDrip posts a JSON-RPC drip request to the faucet: params hold the
// recipient and, optionally, the amount, and the captcha token goes in meta, as
// faucet-hub sends it. It returns the HTTP status and, when the faucet answered
// with JSON-RPC, the decoded response.
func sendDrip(t *testing.T, faucetURL string, params []string, captcha string) (int, spec.BaseJSONResponse) {
	t.Helper()

	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "drip",
		"params":  params,
		"meta":    map[string]string{"captcha": captcha},
	})
	require.NoError(t, err)

	httpClient := http.Client{Timeout: e2eRequestTimeout}
	res, err := httpClient.Post(faucetURL, "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer res.Body.Close()

	var rpcResponse spec.BaseJSONResponse
	if res.StatusCode == http.StatusOK {
		require.NoError(t, json.NewDecoder(res.Body).Decode(&rpcResponse))
	}

	return res.StatusCode, rpcResponse
}
