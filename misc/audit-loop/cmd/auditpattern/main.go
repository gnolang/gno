package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gnolang/gno/misc/audit-loop/internal/auditpattern"
)

func main() {
	gnoBin := flag.String("gno-bin", "", "path to the gno binary; defaults to PATH lookup")
	format := flag.String("format", "markdown", "report format: markdown or json")
	timeout := flag.Duration("timeout", 2*time.Minute, "overall run timeout")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: auditpattern [flags] expected.yaml [expected.yaml...]")
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	ok := true
	var reports []auditpattern.Report
	for _, path := range flag.Args() {
		rec, err := auditpattern.LoadRecord(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		report := auditpattern.Run(ctx, rec, auditpattern.Options{GNOBin: *gnoBin})
		if !report.OK {
			ok = false
		}
		reports = append(reports, report)
	}

	switch *format {
	case "markdown":
		for i, report := range reports {
			if i > 0 {
				fmt.Println()
			}
			fmt.Print(report.Markdown())
		}
	case "json":
		data, err := auditpattern.ReportsJSON(reports)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		fmt.Println(string(data))
	default:
		fmt.Fprintf(os.Stderr, "unknown format %q\n", *format)
		os.Exit(2)
	}

	if !ok {
		os.Exit(1)
	}
}
