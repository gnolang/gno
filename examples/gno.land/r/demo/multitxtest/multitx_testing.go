package multitxtest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/gnovm/tests"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type TxDefinition struct {
	Pkg         *std.MemPackage
	Entrypoint  string
	OrigSend    std.Coins
	ExpectPanic bool
}

func RunTxs(t *testing.T, txs []TxDefinition) {
	var (
		mode    = tests.ImportModeStdlibsOnly
		rootDir = filepath.Join("..", "..", "..", "..", "..")
		stdin   = os.Stdin
		stdout  = os.Stdout
		stderr  = os.Stderr
		store   = tests.TestStore(rootDir, "", stdin, stdout, stderr, mode)
	)
	store.SetStrictGo2GnoMapping(true) // natives must be registered
	gnolang.DisableDebug()             // until main call
	m := tests.TestMachine(store, stdout, "main")
	for i, tx := range txs {
		ctx := m.Context.(stdlibs.ExecContext)
		ctx.OrigSend = tx.OrigSend
		if i > 0 {
			ctx.Height++
			ctx.Timestamp++
		}
		m.Context = ctx
		memPkg := tx.Pkg
		m.RunMemPackage(memPkg, true)
		store.ClearCache()
		m.PreprocessAllFilesAndSaveBlockNodes()
		pv2 := store.GetPackage(tx.Pkg.Path, false)
		if i == 0 {
			m.SetActivePackage(pv2)
		}
		gnolang.EnableDebug()
		defer func() {
			r := recover()
			if tx.ExpectPanic && r == nil {
				t.Fatalf("expected panic in "+tx.Entrypoint+"\n%s\n", r, m.String())
			}
			if !tx.ExpectPanic && r != nil {
				t.Fatalf(tx.Entrypoint+" panic: %v\n%s\n", r, m.String())
			}
		}()
		m.RunStatement(gnolang.S(gnolang.Call(gnolang.X(tx.Entrypoint))))
		m.CheckEmpty()
	}
}
