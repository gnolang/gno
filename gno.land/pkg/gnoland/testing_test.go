package gnoland_test

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/jaekwon/testify/require"
)

func TestNewTestingApp(t *testing.T) {
	app := gnoland.NewTestingApp()
	require.NotNil(t, app)
	defer app.Close()

	var _ abci.Application = app

	require.Equal(t, app.AppVersion(), "dev")
	require.Equal(t, app.LastBlockHeight(), int64(0))
	require.Equal(t, app.Name(), "gnoland")
}

func TestParallelApps(t *testing.T) {
	app1 := gnoland.NewTestingApp()
	require.NotNil(t, app1)
	defer app1.Close()

	app2 := gnoland.NewTestingApp()
	require.NotNil(t, app2)
	defer app2.Close()

	// XXX: more advanced testing
}

func TestAppClose(t *testing.T) {
	app := gnoland.NewTestingApp()
	app.Close()
	t.Skip("not implemented")
}

func TestBasicFlow(t *testing.T) {
	app := gnoland.NewTestingApp()
	defer app.Close()
	require.NotNil(t, app)
	// XXX: continue
}

func ExampleNewTestingApp() {
	app := gnoland.NewTestingApp()
	defer app.Close()

	_, _, addr := tu.KeyTestPubAddr()
	resp := app.Query(abci.RequestQuery{
		Path: fmt.Sprintf("bank/%s/%s", bank.QueryBalance, addr.String()),
		Data: []byte{},
	})
	_ = resp

	// XXX: continue

	// Output:
}
