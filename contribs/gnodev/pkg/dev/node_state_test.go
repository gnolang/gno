package dev

import (
	"context"
	"strconv"
	"testing"
	"time"

	emitter "github.com/gnolang/gno/contribs/gnodev/internal/mock"
	"github.com/gnolang/gno/contribs/gnodev/pkg/events"
	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testCounterRealm = "gno.land/r/dev/counter"

func TestNodeMovePreviousTX(t *testing.T) {
	const callInc = 5

	node, emitter := testingCounterRealm(t, callInc)

	t.Run("Prev TX", func(t *testing.T) {
		ctx := testingContext(t)
		err := node.MoveToPreviousTX(ctx)
		require.NoError(t, err)
		assert.Equal(t, events.EvtReload, emitter.NextEvent().Type())

		// Check for correct render update
		render, err := testingRenderRealm(t, node, testCounterRealm)
		require.NoError(t, err)
		require.Equal(t, render, "4")
	})

	t.Run("Next TX", func(t *testing.T) {
		ctx := testingContext(t)
		err := node.MoveToNextTX(ctx)
		require.NoError(t, err)
		assert.Equal(t, events.EvtReload, emitter.NextEvent().Type())

		// Check for correct render update
		render, err := testingRenderRealm(t, node, testCounterRealm)
		require.NoError(t, err)
		require.Equal(t, render, "5")
	})

	t.Run("Multi Move TX", func(t *testing.T) {
		ctx := testingContext(t)
		moves := []struct {
			Move           int
			ExpectedResult string
		}{
			{-2, "3"},
			{2, "5"},
			{-5, "0"},
			{5, "5"},
			{-100, "0"},
			{100, "5"},
			{0, "5"},
		}

		t.Logf("initial state %d", callInc)
		for _, tc := range moves {
			t.Logf("moving from `%d`", tc.Move)
			err := node.MoveBy(ctx, tc.Move)
			require.NoError(t, err)
			if tc.Move != 0 {
				assert.Equal(t, events.EvtReload, emitter.NextEvent().Type())
			}

			// Check for correct render update
			render, err := testingRenderRealm(t, node, testCounterRealm)
			require.NoError(t, err)
			require.Equal(t, render, tc.ExpectedResult)
		}
	})
}

func TestSaveCurrentState(t *testing.T) {
	ctx := testingContext(t)

	node, emitter := testingCounterRealm(t, 2)

	// Save current state
	err := node.SaveCurrentState(ctx)
	require.NoError(t, err)

	// Send a new tx
	msg := gnoclient.MsgCall{
		PkgPath:  testCounterRealm,
		FuncName: "Inc",
		Args:     []string{"10"},
	}

	res, err := testingCallRealm(t, node, msg)
	require.NoError(t, err)
	require.NoError(t, res.CheckTx.Error)
	require.NoError(t, res.DeliverTx.Error)
	assert.Equal(t, events.EvtTxResult, emitter.NextEvent().Type())

	// Test render
	render, err := testingRenderRealm(t, node, testCounterRealm)
	require.NoError(t, err)
	require.Equal(t, render, "12") // 2 + 10

	// Reset state
	err = node.Reset(ctx)
	require.NoError(t, err)
	assert.Equal(t, events.EvtReset, emitter.NextEvent().Type())

	render, err = testingRenderRealm(t, node, testCounterRealm)
	require.NoError(t, err)
	require.Equal(t, render, "2") // Back to the original state
}

func TestExportState(t *testing.T) {
	node, _ := testingCounterRealm(t, 3)

	t.Run("export state", func(t *testing.T) {
		ctx := testingContext(t)
		state, err := node.ExportCurrentState(ctx)
		require.NoError(t, err)
		assert.Equal(t, 3, len(state))
	})

	t.Run("export genesis doc", func(t *testing.T) {
		ctx := testingContext(t)
		doc, err := node.ExportStateAsGenesis(ctx)
		require.NoError(t, err)
		require.NotNil(t, doc.AppState)

		state, ok := doc.AppState.(gnoland.GnoGenesisState)
		require.True(t, ok)
		assert.Equal(t, 3, len(state.Txs))
	})
}

func testingCounterRealm(t *testing.T, inc int) (*Node, *emitter.ServerEmitter) {
	t.Helper()

	const (
		// foo package
		counterGnoMod = "module gno.land/r/dev/counter\n"
		counterFile   = `package counter
import "strconv"

var value int = 0
func Inc(v int) { value += v } // method to increment value
func Render(_ string) string { return strconv.Itoa(value) }
`
	)

	// Generate package counter
	counterPkg := generateTestingPackage(t,
		"gno.mod", counterGnoMod,
		"foo.gno", counterFile)

	// Call NewDevNode with no package should work
	node, emitter := newTestingDevNode(t, counterPkg)
	assert.Len(t, node.ListPkgs(), 1)

	// Test rendering
	render, err := testingRenderRealm(t, node, testCounterRealm)
	require.NoError(t, err)
	require.Equal(t, render, "0")

	// Increment the counter 10 times
	for i := 0; i < inc; i++ {
		t.Logf("call %d", i)
		// Craft `Inc` msg
		msg := gnoclient.MsgCall{
			PkgPath:  testCounterRealm,
			FuncName: "Inc",
			Args:     []string{"1"},
		}

		res, err := testingCallRealm(t, node, msg)
		require.NoError(t, err)
		require.NoError(t, res.CheckTx.Error)
		require.NoError(t, res.DeliverTx.Error)
		assert.Equal(t, events.EvtTxResult, emitter.NextEvent().Type())
	}

	render, err = testingRenderRealm(t, node, testCounterRealm)
	require.NoError(t, err)
	require.Equal(t, render, strconv.Itoa(inc))

	return node, emitter
}

func testingContext(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*7)
	t.Cleanup(cancel)
	return ctx
}
