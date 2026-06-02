// Command multiarch-determinism embeds gnovm directly: it builds a
// gno.Machine, loads the stdlib, parses an embedded corpus.gno program,
// and runs it. The corpus calls into gno stdlibs (crypto/sha256,
// crypto/ed25519, crypto/chacha20, ...) and prints one canonical line
// per case to stdout via gno's println.
//
// The whole point is to test the *gno* point of view: native bindings
// flow through gnovm's Go2Gno/Gno2Go marshalling, and pure-gno code
// runs through gnovm's interpreter. Cross-compiling this small binary
// for several architectures and diffing the stdouts catches consensus-
// relevant non-determinism wherever it lives in that stack.
package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/std"
)

//go:embed corpus.gno
var corpusSrc string

const maxAllocBytes = 500_000_000 // mirrors gno run's cap

func main() {
	rootDir := os.Getenv("GNOROOT")
	if rootDir == "" {
		guessed, err := gnoenv.GuessRootDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "GNOROOT must be set:", err)
			os.Exit(2)
		}
		rootDir = guessed
	}

	output := test.OutputWithError(os.Stdout, os.Stderr)
	_, store := test.ProdStore(rootDir, output, nil)

	const pkgPath = "main"
	ctx := test.Context("", pkgPath, std.Coins{})
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath:       pkgPath,
		Output:        output,
		Store:         store,
		MaxAllocBytes: maxAllocBytes,
		Context:       ctx,
	})
	defer m.Release()

	fn, err := m.ParseFile("corpus.gno", corpusSrc)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse corpus.gno:", err)
		os.Exit(1)
	}
	m.RunFiles(fn)
	m.RunMain()
}
