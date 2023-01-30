package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/peterbourgon/ff/v3/ffcli"
	"io"
	"net/http"
	"os"
)

func DefaultOpts() Opts {
	return Opts{
		changelog:              true,
		backlog:                true,
		curation:               true,
		tips:                   true,
		format:                 "json",
		From:                   "",
		twitterToken:           "",
		githubToken:            "",
		help:                   false,
		httpClient:             &http.Client{},
		twitterSearchTweetsUrl: "https://api.twitter.com/2/tweets/search/recent?query=%23gnotips&max_results=100",
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

func runMain(args []string) error {
	var root *ffcli.Command
	{
		globalFlags := flag.NewFlagSet("gno-reporting", flag.ExitOnError)
		globalFlags.BoolVar(&opts.changelog, "changelog", opts.changelog, "generate changelog")
		globalFlags.BoolVar(&opts.backlog, "backlog", opts.backlog, "generate backlog")
		globalFlags.BoolVar(&opts.curation, "curation", opts.curation, "generate curation")
		globalFlags.BoolVar(&opts.tips, "tips", opts.tips, "generate tips")
		globalFlags.StringVar(&opts.From, "from", opts.From, "from date")
		globalFlags.StringVar(&opts.twitterToken, "twitter-token", opts.twitterToken, "twitter token")
		globalFlags.StringVar(&opts.githubToken, "github-token", opts.githubToken, "github token")
		globalFlags.BoolVar(&opts.help, "help", false, "show help")
		globalFlags.StringVar(&opts.format, "format", opts.format, "output format")
		root = &ffcli.Command{
			ShortUsage: "reporting [flags]",
			FlagSet:    globalFlags,
			Exec: func(ctx context.Context, args []string) error {
				if opts.help {
					return flag.ErrHelp
				}
				changelog, err := fetchChangelog()
				if err != nil {
					return err
				}
				backlog, err := fetchBacklog()
				if err != nil {
					return err
				}
				curation, err := fetchCuration()
				if err != nil {
					return err
				}
				tips, err := fetchTips()
				if err != nil {
					return err
				}
				fmt.Println(changelog + backlog + curation + tips)
				return nil
			},
		}
	}
	return root.ParseAndRun(context.Background(), args)
}

func fetchChangelog() (string, error) {
	return "", nil
}

func fetchBacklog() (string, error) {
	return "", nil
}

func fetchCuration() (string, error) {
	return "", nil
}

func fetchTips() (string, error) {
	var bearer = "Bearer " + opts.twitterToken
	req, err := http.NewRequest("GET", opts.twitterSearchTweetsUrl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", bearer)
	resp, err := opts.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

type Opts struct {
	changelog              bool
	backlog                bool
	curation               bool
	tips                   bool
	From                   string
	twitterToken           string
	githubToken            string
	format                 string
	help                   bool
	httpClient             *http.Client
	twitterSearchTweetsUrl string
}
