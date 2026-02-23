package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime"

	bm "github.com/gnolang/gno/gnovm/pkg/benchops"
)

var (
	outFlag   = flag.String("out", "results.csv", "the out put file")
	benchFlag = flag.String("bench", "./pkg/benchops/gno", "the path to the benchmark contract")
	binFlag   = flag.String("bin", "", "interpret the existing benchmarking file.")
)

// We dump the benchmark in bytes for speed and minimal overhead.
const tmpFile = "benchmark.bin"

func main() {
	runtime.GOMAXPROCS(1) // for consistent benchmarking
	flag.Parse()
	if *binFlag != "" {
		binFile, err := filepath.Abs(*binFlag)
		if err != nil {
			log.Fatal("unable to get absolute path for the file", err)
		}
		stats(binFile)
		return
	}
	bm.Init(tmpFile)
	bstore := benchmarkDiskStore()
	defer bstore.Delete()

	dir, err := filepath.Abs(*benchFlag)
	if err != nil {
		log.Fatal("unable to get absolute path for storage directory.", err)
	}

	// load  stdlibs
	loadStdlibs(bstore)

	if bm.OpsEnabled {
		benchmarkOpCodes(bstore.gnoStore, dir)
	}
	if bm.StorageEnabled {
		benchmarkStorage(bstore, dir)
	}

	if bm.NativeEnabled {
		benchmarkNative(bstore.gnoStore, dir)
	}
	if bm.PreprocessEnabled {
		benchmarkPreprocess(bstore.gnoStore, dir)
	}

	bm.Finish()
	stats(tmpFile)
	err = os.Remove(tmpFile)
	if err != nil {
		log.Printf("Error removing tmp file: %v", err)
	}
}
