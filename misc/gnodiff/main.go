// Command gnodiff runs a seed corpus through the standard Go toolchain and
// through GnoVM and writes a divergence report.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gnolang/gno/misc/gnodiff/internal/diff"
)

func main() {
	corpus := flag.String("corpus", "", "directory of <name>/prog.src seeds (default: ./corpus)")
	out := flag.String("out", "gnodiff-out", "directory for report.txt, report.json, and repros")
	timeout := flag.Duration("timeout", 20*time.Second, "per-program gc timeout; the same deadline treats a hang as non-termination")
	gas := flag.Int64("gas", 50_000_000, "GnoVM gas limit; exhaustion is non-termination, not a gas diff")
	minimize := flag.Bool("minimize", true, "shrink unexpected divergences before saving a repro")
	budget := flag.Int("budget", 30, "maximum minimization attempts per unexpected divergence")
	flag.Parse()

	dir := *corpus
	if dir == "" {
		dir = diff.CorpusDir()
	}
	progs, err := diff.LoadDir(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Fprintf(os.Stderr, "gnodiff: %d programs from %s\n", len(progs), dir)

	rep := diff.RunCorpus(progs, diff.RunConfig{
		Options: diff.Options{
			Timeout:  *timeout,
			GasLimit: *gas,
		},
		Minimize:       *minimize,
		MinimizeBudget: *budget,
		OnCase: func(item diff.CaseResult) {
			status := "ok"
			switch {
			case item.Go.Kind == diff.KindInternal || item.Gno.Kind == diff.KindInternal:
				status = "RUNNER"
			case !item.Match && item.Known != "":
				status = "known"
			case !item.Match:
				status = "DIFF"
			}
			fmt.Fprintf(os.Stderr, "  %-8s %s\n", status, item.Name)
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
