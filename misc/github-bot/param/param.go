package param

import (
	"flag"
	"fmt"
	"os"

	"github.com/sethvargo/go-githubactions"
)

type Params struct {
	Owner   string
	Repo    string
	PrAll   bool
	PrNums  PrList
	Verbose bool
	DryRun  bool
	Timeout uint
}

// Get Params from both cli flags and/or Github Actions context
func Get() Params {
	p := Params{}

	// Add cmd description to usage message
	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), "This tool checks if requirements for a PR to be merged are met (defined in config.go) and display PR status checks accordingly.\n")
		fmt.Fprint(flag.CommandLine.Output(), "A valid Github Token must be provided by setting the GITHUB_TOKEN env variable.\n\n")
		flag.PrintDefaults()
	}

	// Helper to display an error + usage message before exiting
	errorUsage := func(error string) {
		fmt.Fprintf(flag.CommandLine.Output(), "Error : %s\n\n", error)
		flag.Usage()
		os.Exit(1)
	}

	// Flags definition
	flag.StringVar(&p.Owner, "owner", "", "owner of the repo to check, if empty, will be retrieved from Github Actions context")
	flag.StringVar(&p.Repo, "repo", "", "repo to check, if empty, will be retrieved from Github Actions context")
	flag.BoolVar(&p.PrAll, "pr-all", false, "validate all pull requests opened on the repo")
	flag.TextVar(&p.PrNums, "pr-numbers", PrList(nil), "pull request(s) to validate, must be a comma seperated list of PR numbers, e.g '42,1337,2345'. If empty, PR to check will be retrived from Github Actions context")
	flag.BoolVar(&p.Verbose, "verbose", false, "set logging level to debug")
	flag.BoolVar(&p.DryRun, "dry-run", false, "print if pull request requirements are met without updating PR checks on Github web interface")
	flag.UintVar(&p.Timeout, "timeout", 0, "timeout in milliseconds")
	flag.Parse()

	// If any arg remain after flags processing
	if len(flag.Args()) > 0 {
		errorUsage(fmt.Sprintf("Unknown arg(s) provided : %v", flag.Args()))
	}

	// Check if flags are coherents
	if p.PrAll && len(p.PrNums) != 0 {
		errorUsage("You must specify at most one of '-pr-all' and '-pr-numbers' flags")
	}

	// If one of these values is empty, it must be retrieved
	// from Github Actions context
	if p.Owner == "" || p.Repo == "" || (len(p.PrNums) == 0 && !p.PrAll) {
		actionCtx, err := githubactions.Context()
		if err != nil {
			errorUsage(fmt.Sprintf("Unable to get Github Actions context : %v", err))
		}

		if p.Owner == "" {
			if p.Owner, _ = actionCtx.Repo(); p.Owner == "" {
				errorUsage("Unable to retrieve owner from Github Actions context, you may want to set it using -onwer flag")
			}
		}
		if p.Repo == "" {
			if _, p.Repo = actionCtx.Repo(); p.Repo == "" {
				errorUsage("Unable to retrieve repo from Github Actions context, you may want to set it using -repo flag")
			}
		}
		if len(p.PrNums) == 0 && !p.PrAll {
			const errMsg = "Unable to retrieve pull request number from Github Actions context, you may want to set it using -pr-numbers flag"
			issue, ok := actionCtx.Event["issue"].(map[string]any)
			if !ok {
				errorUsage(errMsg)
			}
			num, ok := issue["number"].(float64)
			if !ok || num <= 0 {
				errorUsage(errMsg)
			}
			p.PrNums = PrList([]int{int(num)})
		}
	}

	return p
}
