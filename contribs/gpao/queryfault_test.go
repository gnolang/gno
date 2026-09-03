package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/integration"
	vm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// answeringClient is an RPC client whose every ABCI query is answered, with the
// given response, by a node that was reached.
type answeringClient struct {
	rpcclient.Client
	resp abci.ResponseQuery
}

func (c answeringClient) ABCIQuery(context.Context, string, []byte) (*ctypes.ResultABCIQuery, error) {
	return &ctypes.ResultABCIQuery{Response: c.resp}, nil
}

// TestNodeAnswersAboutItselfAreNotAMiss pins the line between "nothing is
// stored at this path" and "the node could not answer". Both arrive as a
// query error from a node that was reached; only the first is evidence about
// the import.
func TestNodeAnswersAboutItselfAreNotAMiss(t *testing.T) {
	cases := map[string]struct {
		answer          abci.Error
		wantUnavailable bool
	}{
		"file is not available":            {vm.InvalidFileError{}, false},
		"file is not available, pointer":   {&vm.InvalidFileError{}, false},
		"package is not available":         {vm.InvalidPackageError{}, false},
		"internal error, state unloadable": {std.InternalError{}, true},
		"unknown request, route missing":   {std.UnknownRequestError{}, true},
		"string error":                     {abci.StringError("failed to load state at height 12"), true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			g := newRPCGetter(answeringClient{resp: abci.ResponseQuery{
				ResponseBase: abci.ResponseBase{Error: tc.answer},
			}})

			body, err := g.query("gno.land/p/nobody/nothing")
			require.Nil(t, body)
			require.Error(t, err)
			assert.Equal(t, tc.wantUnavailable, errors.Is(err, errResolverUnavailable),
				"a node saying something about itself is not evidence about the import; a node saying nothing is stored is")
			if tc.wantUnavailable {
				assert.ErrorIs(t, g.transportErr, errResolverUnavailable,
					"the fault must be remembered so a failed typecheck is outranked by it")
			} else {
				assert.NoError(t, g.transportErr,
					"a genuine miss records no fault, so the typecheck's verdict stands")
			}
		})
	}
}

// TestGenuineMissFromARealNode pins the wire form of a miss, so the type
// switch above matches what a node actually sends rather than what the keeper
// constructs.
func TestGenuineMissFromARealNode(t *testing.T) {
	cfg := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
	cfg.SkipGenesisSigVerification = true
	node, remote := integration.TestingInMemoryNode(t, log.NewNoopLogger(), cfg)
	defer node.Stop()

	rpc, err := rpcclient.NewHTTPClient(remote)
	require.NoError(t, err)
	g := newRPCGetter(rpc)

	const path = "gno.land/p/nobody/nothing"
	body, err := g.query(path)
	require.Nil(t, body)
	require.Error(t, err)
	assert.NotErrorIs(t, err, errResolverUnavailable,
		"a healthy node answering that nothing is stored at %q is a verdict, not a fault: %T", path, err)
	assert.NoError(t, g.transportErr)
	assert.Nil(t, g.GetMemPackage(path))
}

// TestStalledRemoteIsNotABudgetOverrun: a node that accepts the connection and
// never answers must be reported as unavailable, not as a package that ran out
// of verification time.
func TestStalledRemoteIsNotABudgetOverrun(t *testing.T) {
	stall := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-stall:
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(srv.Close)
	t.Cleanup(func() { close(stall) })

	o := newTestOracle(t)
	o.cfg.remote = srv.URL
	o.cfg.verifyBudget = 3 * time.Second

	err := o.verify(context.Background(), packageImportingChainOnly())
	require.ErrorIs(t, err, errVerifyUnavailable,
		"the network under the resolver stalled; that says nothing about the package")
	assert.NotErrorIs(t, err, errVerifyBudget,
		"a stall must not count against the package's budget allowance, which gives up after %d", maxOverBudgetAttempts)
}
