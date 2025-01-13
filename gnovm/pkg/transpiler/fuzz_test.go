package transpiler_test

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/transpiler"
)

func FuzzTranspiling(f *testing.F) {
	if testing.Short() {
		f.Skip("Running in -short mode")
	}

	// 1. Derive the seeds from our seedGnoFiles.
	ffs := os.DirFS(filepath.Join(gnoenv.RootDir(), "examples"))
	fs.WalkDir(ffs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			panic(err)
		}
		if !strings.HasSuffix(path, ".gno") {
			return nil
		}
		file, err := ffs.Open(path)
		if err != nil {
			panic(err)
		}
		blob, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			panic(err)
		}
		f.Add(blob)
		return nil
	})

	// 2. Run the fuzzers.
	f.Fuzz(func(t *testing.T, gnoSourceCode []byte) {
		// 3. Add timings to ensure that if transpiling takes a long time
		// to run, that we report this as problematic.
		doneCh := make(chan bool, 1)
		readyCh := make(chan bool)
		go func() {
			defer func() {
				r := recover()
				if r == nil {
					return
				}

				sr := fmt.Sprintf("%s", r)
				if !strings.Contains(sr, "invalid line number ") {
					panic(r)
				}
			}()
			close(readyCh)
			defer close(doneCh)
			_, _ = transpiler.Transpile(string(gnoSourceCode), "gno", "in.gno")
			doneCh <- true
		}()

		<-readyCh

		select {
		case <-time.After(2 * time.Second):
			t.Fatalf("took more than 2 seconds to transpile\n\n%s", gnoSourceCode)
		case <-doneCh:
		}
	})
}
