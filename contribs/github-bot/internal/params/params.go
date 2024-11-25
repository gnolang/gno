package params

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/sethvargo/go-githubactions"
)

type Params struct {
	Owner   string
	Repo    string
	PRAll   bool
	PRNums  PRList
	Verbose bool
	DryRun  bool
	Timeout time.Duration
	flagSet *flag.FlagSet
}

func (p *Params) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&p.Owner,
		"owner",
		"",
		"owner of the repo to process, if empty, will be retrieved from GitHub Actions context",
	)

	fs.StringVar(
		&p.Repo,
		"repo",
		"",
		"repo to process, if empty, will be retrieved from GitHub Actions context",
	)

	fs.BoolVar(
		&p.PRAll,
		"pr-all",
		false,
		"process all opened pull requests",
	)

	fs.TextVar(
		&p.PRNums,
		"pr-numbers",
		PRList(nil),
		"pull request(s) to process, must be a comma separated list of PR numbers, e.g '42,1337,7890'. If empty, will be retrieved from GitHub Actions context",
	)

	fs.BoolVar(
		&p.Verbose,
		"verbose",
		false,
		"set logging level to debug",
	)

	fs.BoolVar(
		&p.DryRun,
		"dry-run",
		false,
		"print if pull request requirements are satisfied without updating anything on GitHub",
	)

	fs.DurationVar(
		&p.Timeout,
		"timeout",
		0,
		"timeout after which the bot execution is interrupted",
	)

	p.flagSet = fs
}

func (p *Params) ValidateFlags() {
	// Helper to display an error + usage message before exiting.
	errorUsage := func(err string) {
		fmt.Fprintf(p.flagSet.Output(), "Error: %s\n\n", err)
		p.flagSet.Usage()
		os.Exit(1)
	}

	// Check if flags are coherent.
	if p.PRAll && len(p.PRNums) != 0 {
		errorUsage("You can specify only one of the '-pr-all' and '-pr-numbers' flags")
	}

	// If one of these values is empty, it must be retrieved
	// from GitHub Actions context.
	if p.Owner == "" || p.Repo == "" || (len(p.PRNums) == 0 && !p.PRAll) {
		actionCtx, err := githubactions.Context()
		if err != nil {
			errorUsage(fmt.Sprintf("Unable to get GitHub Actions context: %v", err))
		}

		if p.Owner == "" {
			if p.Owner, _ = actionCtx.Repo(); p.Owner == "" {
				errorUsage("Unable to retrieve owner from GitHub Actions context, you may want to set it using -onwer flag")
			}
		}
		if p.Repo == "" {
			if _, p.Repo = actionCtx.Repo(); p.Repo == "" {
				errorUsage("Unable to retrieve repo from GitHub Actions context, you may want to set it using -repo flag")
			}
		}
		if len(p.PRNums) == 0 && !p.PRAll {
			const errMsg = "Unable to retrieve pull request number from GitHub Actions context, you may want to set it using -pr-numbers flag"
			var num float64

			switch actionCtx.EventName {
			case "issue_comment":
				issue, ok := actionCtx.Event["issue"].(map[string]any)
				if !ok {
					errorUsage(errMsg)
				}
				num, ok = issue["number"].(float64)
				if !ok || num <= 0 {
					errorUsage(errMsg)
				}
			case "pull_request":
				pr, ok := actionCtx.Event["pull_request"].(map[string]any)
				if !ok {
					errorUsage(errMsg)
				}
				num, ok = pr["number"].(float64)
				if !ok || num <= 0 {
					errorUsage(errMsg)
				}
			default:
				errorUsage(errMsg)
			}

			p.PRNums = PRList([]int{int(num)})
		}
	}
}
