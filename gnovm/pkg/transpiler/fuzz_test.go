package transpiler

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func FuzzTranspiling(f *testing.F) {
	if testing.Short() {
		f.Skip("Running in -short mode")
	}

	// 1. Derive the seeds from our seedGnoFiles.
	breakRoot := filepath.Join("gnolang", "gno")
	pc, thisFile, _, _ := runtime.Caller(0)
	index := strings.Index(thisFile, breakRoot)
	_ = pc // to silence the pedantic golangci linter.
	rootPath := thisFile[:index+len(breakRoot)]
	examplesDir := filepath.Join(rootPath, "examples")
	ffs := os.DirFS(examplesDir)
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
			_, _ = Transpile(string(gnoSourceCode), "gno", "in.gno")
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
