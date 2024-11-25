package conditions

import (
	"fmt"
	"regexp"

	"github.com/gnolang/gno/contribs/github-bot/internal/client"
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

// FileChanged Condition.
type fileChanged struct {
	gh      *client.GitHub
	pattern *regexp.Regexp
}

var _ Condition = &fileChanged{}

func (fc *fileChanged) IsMet(pr *github.PullRequest, details treeprint.Tree) bool {
	detail := fmt.Sprintf("A changed file matches this pattern: %s", fc.pattern.String())
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
			fc.gh.Logger.Errorf("Unable to list changed files for PR %d: %v", pr.GetNumber(), err)
			break
		}

		for _, file := range files {
			if fc.pattern.MatchString(file.GetFilename()) {
				return utils.AddStatusNode(true, fmt.Sprintf("%s (filename: %s)", detail, file.GetFilename()), details)
			}
		}

		if response.NextPage == 0 {
			break
		}
		opts.Page = response.NextPage
	}

	return utils.AddStatusNode(false, detail, details)
}

func FileChanged(gh *client.GitHub, pattern string) Condition {
	return &fileChanged{gh: gh, pattern: regexp.MustCompile(pattern)}
}
