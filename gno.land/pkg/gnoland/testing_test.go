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
	println(app)
}

func ExampleNewTestingApp() {
	_, _, addr := tu.KeyTestPubAddr()
	app := gnoland.NewTestingApp()
	fmt.Println("app", app)
	resp := app.Query(abci.RequestQuery{
		Path: fmt.Sprintf("bank/%s/%s", bank.QueryBalance, addr.String()),
		Data: []byte{},
	})
	fmt.Println("resp", resp)
	// Output:
	// ...
}
