package condition

import (
	"bot/client"
	"fmt"
	"regexp"

	"github.com/google/go-github/v66/github"
)

// FileChanged Condition
type fileChanged struct {
	gh      *client.Github
	pattern *regexp.Regexp
}

var _ Condition = &fileChanged{}

// Validate implements Condition
func (fc *fileChanged) Validate(pr *github.PullRequest) bool {
	opts := &github.ListOptions{
		PerPage: client.PageSize,
	}

	for {
		files, response, err := fc.gh.Client.PullRequests.ListFiles(
			fc.gh.Ctx,
			fc.gh.Owner,
			fc.gh.Repo,
			pr.GetNumber(),
			opts,
		)
		if err != nil {
			fc.gh.Logger.Errorf("Unable to list changed files for PR %d : %v", pr.GetNumber(), err)
			break
		}

		for _, file := range files {
			if fc.pattern.MatchString(file.GetFilename()) {
				fc.gh.Logger.Debugf("File %s is matching pattern %s", file.GetFilename(), fc.pattern.String())
				return true
			}
		}

		if response.NextPage == 0 {
			break
		}
		opts.Page = response.NextPage
	}

	return false
}

// GetText implements Condition
func (fc *fileChanged) GetText() string {
	return fmt.Sprintf("A changed file match this pattern : %s", fc.pattern.String())
}

func FileChanged(gh *client.Github, pattern string) Condition {
	return &fileChanged{gh: gh, pattern: regexp.MustCompile(pattern)}
}
