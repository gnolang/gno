package node

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	cfg "github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
)

// App that refuses any genesis whose InitialHeight is not wantHeight.
type pickyApp struct {
	abci.BaseApplication
	wantHeight int64
	seen       []int64
}

func (a *pickyApp) InitChain(req abci.RequestInitChain) abci.ResponseInitChain {
	a.seen = append(a.seen, req.InitialHeight)
	if req.InitialHeight != a.wantHeight {
		return abci.ResponseInitChain{ResponseBase: abci.ResponseBase{
			Error: abci.StringError("InitialHeight mismatch: refusing to boot"),
		}}
	}
	return abci.ResponseInitChain{Validators: req.Validators}
}

// The end-to-end path the reset exists for: the app refuses the genesis through
// ResponseInitChain.Error, the operator corrects the field the rejection named,
// and the restart against the same data dir must boot.
func TestE2ERejectThenCorrectedGenesisBoots(t *testing.T) {
	config, genesisFile := cfg.ResetTestRoot("node_e2e_reject_test")
	defer os.RemoveAll(config.RootDir)

	app := &pickyApp{wantHeight: 1}
	config.LocalApp = app

	build := func(initialHeight int64) (*Node, error) {
		provider := func() (*types.GenesisDoc, error) {
			doc, derr := types.GenesisDocFromFile(genesisFile)
			if derr != nil {
				return nil, derr
			}
			doc.InitialHeight = initialHeight
			doc.AppState = "app-state"
			return doc, nil
		}
		return DefaultNewNodeWithGenesisProvider(
			config, provider, events.NewEventSwitch(), log.NewNoopLogger())
	}

	// genesis.json says 42; the app wants 1.
	_, err := build(42)
	require.Error(t, err, "a rejected genesis must not boot")
	t.Logf("boot1 err=%v", err)

	// Operator corrects the file and restarts against the same data dir.
	n, err := build(1)
	require.NoError(t, err, "the corrected genesis must boot")
	require.NotNil(t, n)
	assert.Equal(t, int64(1), n.GenesisDoc().InitialHeight, "corrected doc must be the one in use")
	assert.Equal(t, []int64{42, 1}, app.seen, "the app must see the corrected InitialHeight on the retry")
}
