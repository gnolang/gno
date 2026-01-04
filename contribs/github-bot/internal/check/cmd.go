package check

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gnolang/gno/contribs/github-bot/internal/utils"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/sethvargo/go-githubactions"
)

type checkFlags struct {
	Owner   string
	Repo    string
	PRAll   bool
	PRNums  utils.PRList
	Verbose *bool
	DryRun  bool
	Timeout time.Duration
	flagSet *flag.FlagSet
}

func NewCheckCmd(verbose *bool) *commands.Command {
	flags := &checkFlags{Verbose: verbose}

	return commands.NewCommand(
		commands.Metadata{
			Name:       "check",
			ShortUsage: "github-bot check [flags]",
			ShortHelp:  "checks requirements for a pull request to be merged",
			LongHelp:   "This tool checks if the requirements for a pull request to be merged are satisfied (defined in ./internal/config/config.go) and displays PR status checks accordingly.\nA valid GitHub Token must be provided by setting the GITHUB_TOKEN environment variable.",
		},
		flags,
		func(_ context.Context, _ []string) error {
			flags.validateFlags()
			return execCheck(flags)
		},
	)
}

func (flags *checkFlags) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&flags.Owner,
		"owner",
		"",
		"owner of the repo to process, if empty, will be retrieved from GitHub Actions context",
	)

	fs.StringVar(
		&flags.Repo,
		"repo",
		"",
		"repo to process, if empty, will be retrieved from GitHub Actions context",
	)

	fs.BoolVar(
		&flags.PRAll,
		"pr-all",
		false,
		"process all opened pull requests",
	)

	fs.TextVar(
		&flags.PRNums,
		"pr-numbers",
		utils.PRList(nil),
		"pull request(s) to process, must be a comma separated list of PR numbers, e.g '42,1337,7890'. If empty, will be retrieved from GitHub Actions context",
	)

	fs.BoolVar(
		&flags.DryRun,
		"dry-run",
		false,
		"print if pull request requirements are satisfied without updating anything on GitHub",
	)

	fs.DurationVar(
		&flags.Timeout,
		"timeout",
		0,
		"timeout after which the bot execution is interrupted",
	)

	flags.flagSet = fs
}

func (flags *checkFlags) validateFlags() {
	// Helper to display an error + usage message before exiting.
	errorUsage := func(err string) {
		fmt.Fprintf(flags.flagSet.Output(), "Error: %s\n\n", err)
		flags.flagSet.Usage()
		os.Exit(1)
	}

	// Check if flags are coherent.
	if flags.PRAll && len(flags.PRNums) != 0 {
		errorUsage("You can specify only one of the '-pr-all' and '-pr-numbers' flags.")
	}

	// If one of these values is empty, it must be retrieved
	// from GitHub Actions context.
	if flags.Owner == "" || flags.Repo == "" || (len(flags.PRNums) == 0 && !flags.PRAll) {
		actionCtx, err := githubactions.Context()
		if err != nil {
			errorUsage(fmt.Sprintf("Unable to get GitHub Actions context: %v.", err))
		}

		if flags.Owner == "" {
			if flags.Owner, _ = actionCtx.Repo(); flags.Owner == "" {
				errorUsage("Unable to retrieve owner from GitHub Actions context, you may want to set it using -onwer flag.")
			}
		}
		if flags.Repo == "" {
			if _, flags.Repo = actionCtx.Repo(); flags.Repo == "" {
				errorUsage("Unable to retrieve repo from GitHub Actions context, you may want to set it using -repo flag.")
			}
		}

		if len(flags.PRNums) == 0 && !flags.PRAll {
			prNum, err := utils.GetPRNumFromActionsCtx(actionCtx)
			if err != nil {
				errorUsage(fmt.Sprintf("Unable to retrieve pull request number from GitHub Actions context: %s\nYou may want to set it using -pr-numbers flag.", err.Error()))
			}

			flags.PRNums = utils.PRList{prNum}
		}
	}
}
