// Command gnodiff runs a seed corpus through the standard Go toolchain and
// through GnoVM and writes a divergence report.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gnolang/gno/misc/gnodiff/internal/diff"
	"github.com/gnolang/gno/misc/gnodiff/internal/gen"
)

func main() {
	corpus := flag.String("corpus", "", "directory of <name>/prog.src seeds (default: ./corpus)")
	out := flag.String("out", "gnodiff-out", "directory for report.txt, report.json, and repros")
	timeout := flag.Duration("timeout", 20*time.Second, "per-program gc timeout; the same deadline treats a hang as non-termination")
	gas := flag.Int64("gas", 50_000_000, "GnoVM gas limit; exhaustion is non-termination, not a gas diff")
	minimize := flag.Bool("minimize", true, "shrink real divergences before saving a repro")
	budget := flag.Int("budget", 30, "maximum minimization attempts per real divergence")
	writeCorpus := flag.String("write-corpus", "", "write the enumerated seed catalog into this directory and exit")
	generate := flag.Int("generate", 0, "run this many generated programs instead of the seed corpus")
	seed := flag.Uint64("seed", 1, "first seed for -generate")
	flag.Parse()

	if *generate < 0 {
		fmt.Fprintln(os.Stderr, "-generate must be >= 0")
		os.Exit(2)
	}
	if *generate > 0 && *corpus != "" {
		fmt.Fprintln(os.Stderr, "-generate and -corpus are mutually exclusive")
		os.Exit(2)
	}
	if *timeout <= 0 {
		fmt.Fprintln(os.Stderr, "-timeout must be > 0")
		os.Exit(2)
	}
	if *gas <= 0 {
		fmt.Fprintln(os.Stderr, "-gas must be > 0")
		os.Exit(2)
	}
	if *budget < 0 {
		fmt.Fprintln(os.Stderr, "-budget must be >= 0")
		os.Exit(2)
	}

	if *writeCorpus != "" {
		if err := gen.WriteCatalog(*writeCorpus); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		fmt.Fprintf(os.Stderr, "wrote catalog to %s\n", *writeCorpus)
		return
	}

	var progs []diff.Program
	var err error
	dir := *corpus
	if *generate > 0 {
		progs = make([]diff.Program, 0, *generate)
		for i := range *generate {
			s := *seed + uint64(i)*0x9e3779b97f4a7c15
			progs = append(progs, diff.Program{
				Name:   fmt.Sprintf("gen/%d", s),
				Source: gen.Generate(s),
			})
		}
		fmt.Fprintf(os.Stderr, "gnodiff: %d generated programs from seed %d\n", len(progs), *seed)
	} else {
		if dir == "" {
			dir = diff.CorpusDir()
			progs, err = diff.LoadCorpus()
		} else {
			progs, err = diff.LoadDir(dir)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		fmt.Fprintf(os.Stderr, "gnodiff: %d programs from %s\n", len(progs), dir)
	}

	rep := diff.RunCorpus(progs, diff.RunConfig{
		Options: diff.Options{
			Timeout:  *timeout,
			GasLimit: *gas,
		},
		Minimize:       *minimize,
		MinimizeBudget: *budget,
		OnCase: func(item diff.CaseResult) {
			if item.Class == diff.ClassExact {
				return
			}
			fmt.Fprintf(os.Stderr, "  %-14s %s\n", item.Class, item.Name)
		},
	})
	if err := rep.Write(*out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Print(rep.Text())
	fmt.Fprintf(os.Stderr, "report written to %s\n", *out)
	if rep.Failed() {
		os.Exit(1)
	}
}
