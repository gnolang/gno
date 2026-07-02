package healthcheck

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"time"

	rpcClient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type healthcheckCfg struct {
	remote  string
	timeout time.Duration
}

func NewHealthCheckCmd(io commands.IO) *commands.Command {
	cfg := &healthcheckCfg{}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "healthcheck",
			ShortUsage: "healthcheck [flags]",
			ShortHelp:  "run health checks against a running gnoland node",
			LongHelp: `Connects to a running gnoland node and runs a series of health checks
to verify that the chain is functional after an upgrade. Checks include:

  - Node connectivity and status
  - Block production (height is advancing)
  - Basic ABCI query (realm render)

Exit codes:
  0  all checks passed
  1  one or more checks failed`,
		},
		cfg,
		func(ctx context.Context, args []string) error {
			return execHealthCheck(ctx, cfg, io)
		},
	)
}

func (c *healthcheckCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.remote,
		"remote",
		"http://127.0.0.1:26657",
		"RPC address of the gnoland node",
	)

	fs.DurationVar(
		&c.timeout,
		"timeout",
		30*time.Second,
		"timeout for each health check",
	)
}

type checkResult struct {
	name   string
	passed bool
	detail string
	err    error
}

func execHealthCheck(ctx context.Context, cfg *healthcheckCfg, io commands.IO) error {
	io.Printfln("=== Chain Health Check ===")
	io.Printfln("Remote: %s", cfg.remote)
	io.Printfln("")

	client, err := rpcClient.NewHTTPClient(cfg.remote)
	if err != nil {
		return fmt.Errorf("failed to create RPC client: %w", err)
	}

	checks := []checkResult{
		checkStatus(ctx, client),
		checkBlockProduction(ctx, client, cfg.timeout),
		checkABCIQuery(ctx, client),
	}

	var failed int
	for _, c := range checks {
		status := "PASS"
		if !c.passed {
			status = "FAIL"
			failed++
		}
		io.Printfln("[%s] %s", status, c.name)
		if c.detail != "" {
			io.Printfln("       %s", c.detail)
		}
		if c.err != nil {
			io.Printfln("       error: %v", c.err)
		}
	}

	io.Printfln("")
	if failed > 0 {
		io.Printfln("=== %d/%d checks FAILED ===", failed, len(checks))
		return fmt.Errorf("%d health checks failed", failed)
	}

	io.Printfln("=== All %d checks PASSED ===", len(checks))
	return nil
}

func checkStatus(ctx context.Context, client *rpcClient.RPCClient) checkResult {
	res := checkResult{name: "Node status"}

	status, err := client.Status(ctx, nil)
	if err != nil {
		res.err = err
		return res
	}

	res.passed = true
	res.detail = fmt.Sprintf("chain_id=%s height=%d",
		status.NodeInfo.Network,
		status.SyncInfo.LatestBlockHeight,
	)
	return res
}

func checkBlockProduction(ctx context.Context, client *rpcClient.RPCClient, timeout time.Duration) checkResult {
	res := checkResult{name: "Block production"}

	status1, err := client.Status(ctx, nil)
	if err != nil {
		res.err = err
		return res
	}
	h1 := status1.SyncInfo.LatestBlockHeight

	// Wait a bit for a new block
	deadline := time.After(timeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			res.err = errors.New("timed out waiting for new block")
			return res
		case <-ticker.C:
			status2, err := client.Status(ctx, nil)
			if err != nil {
				continue
			}
			h2 := status2.SyncInfo.LatestBlockHeight
			if h2 > h1 {
				res.passed = true
				res.detail = fmt.Sprintf("height advanced %d -> %d", h1, h2)
				return res
			}
		}
	}
}

func checkABCIQuery(ctx context.Context, client *rpcClient.RPCClient) checkResult {
	res := checkResult{name: "ABCI query (vm/qrender)"}

	// Try to render r/sys/params — a basic realm query
	qres, err := client.ABCIQuery(ctx, "vm/qrender", []byte("gno.land/r/sys/params\n"))
	if err != nil {
		res.err = err
		return res
	}

	if qres.Response.IsErr() {
		res.err = fmt.Errorf("query returned error: %s", qres.Response.Log)
		return res
	}

	res.passed = true
	dataLen := len(qres.Response.Data)
	if dataLen > 100 {
		res.detail = fmt.Sprintf("response length=%d bytes (truncated)", dataLen)
	} else {
		res.detail = fmt.Sprintf("response=%q", string(qres.Response.Data))
	}
	return res
}
