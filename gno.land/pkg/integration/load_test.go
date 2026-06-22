package integration

import (
	"testing"

	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// TestExamplesLoad boots an in-memory gnoland node and reports any
// AddPackage tx that fails at genesis. All failures are collected so a
// single run surfaces every load issue.
func TestExamplesLoad(t *testing.T) {
	rootdir := gnoenv.RootDir()

	logger := log.NewTestingLogger(t)
	config, _ := TestingNodeConfig(t, rootdir)
	config.InitChainerConfig.GenesisTxResultHandler = func(_ sdk.Context, tx std.Tx, res sdk.Result) {
		if !res.IsErr() {
			return
		}

		for _, msg := range tx.Msgs {
			if addPkg, ok := msg.(vmm.MsgAddPackage); ok && addPkg.Package != nil {
				t.Errorf("AddPackage failed at genesis: %s: %s", addPkg.Package.Path, res.Log)
				return
			}
		}

		t.Errorf("Msgs failed at genesis: %s", res.Log)
	}

	node, _ := TestingInMemoryNode(t, logger, config)
	t.Cleanup(func() {
		if err := node.Stop(); err != nil {
			t.Logf("node.Stop: %v", err)
		}
	})
}
