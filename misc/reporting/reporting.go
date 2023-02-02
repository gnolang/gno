package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/go-github/v50/github"
	"net/http"
	"os"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"
)

func DefaultOpts() Opts {
	return Opts{
		changelog:              true,
		backlog:                true,
		curation:               true,
		tips:                   true,
		format:                 "json",
		since:                  "",
		twitterToken:           "",
		githubToken:            "",
		help:                   false,
		httpClient:             &http.Client{},
		twitterSearchTweetsUrl: "https://api.twitter.com/2/tweets/search/recent?query=%23gnotips&tweet.fields=created_at&max_results=100",
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
		globalFlags.StringVar(&opts.since, "since", opts.since, "since date RFC 3339 (ex: 2003-01-19T00:00:00Z)")
		globalFlags.StringVar(&opts.twitterToken, "twitter-token", opts.twitterToken, "twitter token")
		globalFlags.StringVar(&opts.githubToken, "github-token", opts.githubToken, "github token")
		globalFlags.BoolVar(&opts.help, "help", false, "show help")
		globalFlags.StringVar(&opts.format, "format", opts.format, "output format")
		globalFlags.StringVar(&opts.outputPath, "output-path", opts.outputPath, "output directory path")
		root = &ffcli.Command{
			ShortUsage: "reporting [flags]",
			FlagSet:    globalFlags,
			Exec: func(ctx context.Context, args []string) error {
				var err error
				since := time.Time{}
				if opts.help {
					return flag.ErrHelp
				}
				if opts.twitterToken == "" && opts.tips {
					return fmt.Errorf("twitter token is required to fetch tips")
				}
				if opts.githubToken == "" && (opts.curation || opts.backlog || opts.changelog) {
					return fmt.Errorf("github token is required to fetch curation, backlog or changelog")
				}
				if opts.since != "" {
					since, err = time.Parse("2006-01-02T15:04:05.000Z", opts.since)
					if err != nil {
						return err
					}
					if err != nil {
						return fmt.Errorf("invalid from date")
					}
				}
				githubClient := initGithubClient()
				err = fetchChangelog(githubClient, since)
				if err != nil {
					return err
				}
				err = fetchBacklog(githubClient, since)
				if err != nil {
					return err
				}
				err = fetchCuration(githubClient, since)
				if err != nil {
					return err
				}
				err = fetchTips()
				if err != nil {
					return err
				}
				return nil
			},
		}
	}
	return root.ParseAndRun(context.Background(), args)
}

// TODO: Fetch changelog recent contributors, new PR merged, new issues closed ...
func fetchChangelog(client *github.Client, since time.Time) error {
	if !opts.changelog {
		return nil
	}

	issues, err := githubFetchIssues(client, &github.IssueListByRepoOptions{State: "closed", Since: since}, "gnolang", "gno")
	if err != nil {
		return err
	}
	_, err = json.Marshal(issues)
	if err != nil {
		return err
	}
	return nil
}

// TODO: Fetch backlog from github issues & PRS ...
func fetchBacklog(client *github.Client, since time.Time) error {
	if !opts.backlog {
		return nil
	}
	issues, err := githubFetchIssues(client, &github.IssueListByRepoOptions{State: "open", Since: since}, "gnolang", "gno")
	if err != nil {
		return err
	}
	_, err = json.Marshal(issues)
	if err != nil {
		return err
	}
	return nil
}

// TODO: Fetch curation from github commits & issues & PRS in `awesome-gno` repo
func fetchCuration(client *github.Client, since time.Time) error {
	if !opts.curation {
		return nil
	}
	issues, err := githubFetchIssues(client, &github.IssueListByRepoOptions{State: "all", Since: since}, "gnolang", "awesome-gno")
	if err != nil {
		return err
	}
	_, err = json.Marshal(issues)
	if err != nil {
		return err
	}
	return nil
}

func fetchTips() error {
	if !opts.tips {
		return nil
	}
	ret := twitterFetchTips()
	err := writeTips(ret)
	if err != nil {
		return err
	}
	return nil
}

type Opts struct {
	changelog              bool
	backlog                bool
	curation               bool
	tips                   bool
	since                  string
	twitterToken           string
	githubToken            string
	format                 string
	help                   bool
	httpClient             *http.Client
	twitterSearchTweetsUrl string
	awesomeGnoRepoUrl      string
	outputPath             string
}
