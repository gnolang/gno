//go:build e2e

// Package e2e runs the whole system against a local gnodev and live drand:
// a user flips a coin, the coinflip realm requests randomness from r/drand,
// the relayer fetches the real future evmnet beacon over HTTP and submits
// it, r/drand verifies it on-chain, and the flip settles.
//
//	go test -tags e2e -v ./e2e
//
// Needs gnodev on PATH (or $GNODEV), built from the same checkout as
// GNOROOT, and network access to the drand HTTP API.
package e2e

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/std"

	"github.com/gnolang/gno/contribs/gnodrand/internal/relayer"
)

const (
	drandPath    = "gno.land/r/drand/v0"
	coinflipPath = "gno.land/r/drand/coinflip/v0"
)

func TestEndToEnd(t *testing.T) {
	signer := newSigner(t)
	info, err := signer.Info()
	if err != nil {
		t.Fatal(err)
	}
	addr := info.GetAddress()

	remote := startGnodev(t, addr.String())
	rpc, err := rpcclient.NewHTTPClient(remote)
	if err != nil {
		t.Fatal(err)
	}
	client := &gnoclient.Client{Signer: signer, RPCClient: rpc}
	waitReady(t, client)

	call := func(pkg, fn string, args ...string) string {
		t.Helper()
		res, err := client.Call(gnoclient.BaseTxCfg{GasFee: "1000000ugnot", GasWanted: 100_000_000},
			vm.MsgCall{Caller: addr, PkgPath: pkg, Func: fn, Args: args})
		if err != nil {
			t.Fatalf("%s.%s: %v", pkg, fn, err)
		}
		return string(res.DeliverTx.Data)
	}

	// A user requests directly, and a realm requests on behalf of a player.
	reqID := unquote(t, call(drandPath, "Request"))
	flipID := unquote(t, call(coinflipPath, "Flip", "true"))
	t.Logf("request %s, flip %s", reqID, flipID)

	pending := qeval(t, client, drandPath, "PendingRounds(10)")
	if pending == `("" string)` {
		t.Fatal("no pending round after Request")
	}
	t.Logf("pending: %s", pending)

	r := &relayer.Relayer{
		Source: relayer.HTTPSource{Mirrors: relayer.DefaultMirrors, Client: &http.Client{Timeout: 5 * time.Second}},
		Chain: relayer.GnoChain{
			Client: client, PkgPath: drandPath,
			FeePerBeacon: std.NewCoin("ugnot", 22_000), GasPerBeacon: 22_000_000,
		},
		Log:      slog.New(slog.NewTextHandler(os.Stderr, nil)),
		MaxBatch: 10,
	}

	deadline := time.Now().Add(60 * time.Second)
	for {
		if _, err := r.Tick(context.Background()); err != nil {
			t.Fatalf("relayer: %v", err)
		}
		if qeval(t, client, drandPath, "PendingRounds(10)") == `("" string)` {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("rounds still pending after 60s")
		}
		time.Sleep(time.Second)
	}

	status := qeval(t, client, drandPath, fmt.Sprintf("Status(%q)", reqID))
	if !strings.Contains(status, "(true bool)") {
		t.Fatalf("request not ready: %s", status)
	}
	call(coinflipPath, "Settle", flipID)
	result := qeval(t, client, coinflipPath, fmt.Sprintf("Result(%q)", flipID))
	if !strings.HasPrefix(strings.TrimSpace(result), "(true bool)") {
		t.Fatalf("flip not settled: %s", result)
	}
	t.Logf("flip result: %s", strings.ReplaceAll(result, "\n", " "))
	t.Logf("drand status:\n%s", render(t, client, drandPath, ""))
}

// newSigner makes a throwaway key; gnodev funds it at genesis.
func newSigner(t *testing.T) gnoclient.Signer {
	t.Helper()
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		t.Fatal(err)
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		t.Fatal(err)
	}
	s, err := gnoclient.SignerFromBip39(mnemonic, "dev", "", 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func startGnodev(t *testing.T, premine string) string {
	t.Helper()
	bin := os.Getenv("GNODEV")
	if bin == "" {
		bin = "gnodev"
	}
	listen := freeAddr(t)
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, bin, "local",
		"-no-web", "-no-watch",
		"-node-rpc-listener", listen,
		"-add-account", premine+"=1000000000000ugnot",
		"-paths", drandPath+","+coinflipPath,
	)
	logf, err := os.Create(filepath.Join(t.TempDir(), "gnodev.log"))
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stdout, cmd.Stderr = logf, logf
	if err := cmd.Start(); err != nil {
		t.Skipf("gnodev not available: %v", err)
	}
	t.Cleanup(func() {
		cancel()
		_ = cmd.Wait()
		if t.Failed() {
			_, _ = logf.Seek(0, io.SeekStart)
			b, _ := io.ReadAll(logf)
			t.Logf("gnodev log:\n%s", b)
		}
		_ = logf.Close()
	})
	return "http://" + listen
}

func freeAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()
	return l.Addr().String()
}

func waitReady(t *testing.T, c *gnoclient.Client) {
	t.Helper()
	deadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		if _, _, err := c.QEval(drandPath, "PendingRounds(1)"); err == nil {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatal("gnodev did not become ready")
}

func qeval(t *testing.T, c *gnoclient.Client, pkg, expr string) string {
	t.Helper()
	res, _, err := c.QEval(pkg, expr)
	if err != nil {
		t.Fatalf("qeval %s: %v", expr, err)
	}
	return strings.TrimSpace(res)
}

func render(t *testing.T, c *gnoclient.Client, pkg, path string) string {
	t.Helper()
	res, _, err := c.Render(pkg, path)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return res
}

var quoted = regexp.MustCompile(`^\("([^"]*)" string\)`)

func unquote(t *testing.T, data string) string {
	t.Helper()
	m := quoted.FindStringSubmatch(strings.TrimSpace(data))
	if m == nil {
		t.Fatalf("unexpected return value %q", data)
	}
	return m[1]
}
