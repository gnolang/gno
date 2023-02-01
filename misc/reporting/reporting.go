package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/go-github/v50/github"
	"net/http"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func DefaultOpts() Opts {
	return Opts{
		changelog:              true,
		backlog:                true,
		curation:               true,
		tips:                   true,
		format:                 "json",
		from:                   "",
		twitterToken:           "",
		githubToken:            "",
		help:                   false,
		httpClient:             &http.Client{},
		twitterSearchTweetsUrl: "https://api.twitter.com/2/tweets/search/recent?query=%23gnotips&max_results=100",
		awesomeGnoRepoUrl:      "https://api.github.com/repos/gnolang/awesome-gno/issues",
		outputPath:             "./output/",
	}
}

var opts = DefaultOpts()

func main() {
	err := runMain(os.Args[1:])
	if err == flag.ErrHelp {
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		os.Exit(1)
	}
}

// TODO: Verify if we can use a template engine to format the output: Like prebuild a report template with a function to format the data
func runMain(args []string) error {
	var root *ffcli.Command
	{
		globalFlags := flag.NewFlagSet("gno-reporting", flag.ExitOnError)
		globalFlags.BoolVar(&opts.changelog, "changelog", opts.changelog, "generate changelog")
		globalFlags.BoolVar(&opts.backlog, "backlog", opts.backlog, "generate backlog")
		globalFlags.BoolVar(&opts.curation, "curation", opts.curation, "generate curation")
		globalFlags.BoolVar(&opts.tips, "tips", opts.tips, "generate tips")
		globalFlags.StringVar(&opts.from, "from", opts.from, "from date")
		globalFlags.StringVar(&opts.twitterToken, "twitter-token", opts.twitterToken, "twitter token")
		globalFlags.StringVar(&opts.twitterToken, "twitter-since", opts.twitterToken, "twitter since date RFC 3339 (ex: 2003-01-19T00:00:00Z)")
		globalFlags.StringVar(&opts.githubToken, "github-token", opts.githubToken, "github token")
		globalFlags.BoolVar(&opts.help, "help", false, "show help")
		globalFlags.StringVar(&opts.format, "format", opts.format, "output format")
		globalFlags.StringVar(&opts.outputPath, "output-path", opts.outputPath, "output directory path")
		root = &ffcli.Command{
			ShortUsage: "reporting [flags]",
			FlagSet:    globalFlags,
			Exec: func(ctx context.Context, args []string) error {

				if opts.help {
					return flag.ErrHelp
				}
				if opts.twitterToken == "" && opts.tips {
					return fmt.Errorf("twitter token is required to fetch tips")
				}
				if opts.githubToken == "" && (opts.curation || opts.backlog || opts.changelog) {
					return fmt.Errorf("github token is required to fetch curation, backlog or changelog")
				}
				githubClient := initGithubClient()
				outputs := map[string]string{
					"changelog": fetchChangelog(githubClient),
					"backlog":   fetchBacklog(githubClient),
					"curation":  fetchCuration(githubClient),
					"tips":      fetchTips(),
				}

				err := writeOutputFiles(outputs)
				if err != nil {
					return err
				}
				return nil
			},
		}
	}
	return root.ParseAndRun(context.Background(), args)
}

// TODO: Fetch changelog recent contributors, new PR merged, new issues closed ... & use from option
func fetchChangelog(client *github.Client) string {
	if !opts.changelog {
		return ""
	}
	// Return a JSON which contains the following data:
	// - contributors (github) (https://api.github.com/repos/gnolang/gno/contributors) (from)
	// - PRs merged (github) (https://api.github.com/repos/gnolang/gno/pulls?state=closed) (from) with issues linked
	// - new releases (github) (https://api.github.com/repos/gnolang/gno/releases) (from)
	return ""
}

// TODO: Fetch backlog from github issues & PRS ... & use from option
func fetchBacklog(client *github.Client) string {
	if !opts.backlog {
		return ""
	}
	// Return a JSON which contains the following data:
	// - new issues (github) (https://api.github.com/repos/gnolang/gno/issues) (from)
	// - new & updated PRs (github) (https://api.github.com/repos/gnolang/gno/pulls) (from)
	return ""
}

// TODO: Fetch curation from github commits & issues & PRS in `awesome-gno` repo & use from option
func fetchCuration(client *github.Client) string {
	if !opts.curation {
		return ""
	}
	issues, err := githubFetchIssues(client, &github.IssueListByRepoOptions{}, "gnolang", "awesome-gno")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err)
		return ""
	}
	for _, issue := range issues {
		fmt.Printf("%+s \n", *issue.Title)
	}
	return "string(body)"
}

func fetchTips() string {
	if !opts.tips {
		return ""
	}
	ret := twitterFetchTips()
	return ret
}

type Opts struct {
	changelog              bool
	backlog                bool
	curation               bool
	tips                   bool
	from                   string
	twitterToken           string
	githubToken            string
	format                 string
	help                   bool
	httpClient             *http.Client
	twitterSearchTweetsUrl string
	awesomeGnoRepoUrl      string
	outputPath             string
}
