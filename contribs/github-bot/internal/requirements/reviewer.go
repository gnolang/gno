package requirements

import (
	"fmt"

	"github.com/gnolang/gno/contribs/github-bot/internal/client"
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

// Reviewer Requirement.
type reviewByUser struct {
	gh   *client.GitHub
	user string
}

var _ Requirement = &reviewByUser{}

func (r *reviewByUser) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	detail := fmt.Sprintf("This user approved pull request: %s", r.user)

	// If not a dry run, make the user a reviewer if he's not already.
	if !r.gh.DryRun {
		requested := false
		reviewers, err := r.gh.ListPRReviewers(pr.GetNumber())
		if err != nil {
			r.gh.Logger.Errorf("unable to check if user %s review is already requested: %v", r.user, err)
			return utils.AddStatusNode(false, detail, details)
		}

		for _, user := range reviewers.Users {
			if user.GetLogin() == r.user {
				requested = true
				break
			}
		}

		if requested {
			r.gh.Logger.Debugf("Review of user %s already requested on PR %d", r.user, pr.GetNumber())
		} else {
			r.gh.Logger.Debugf("Requesting review from user %s on PR %d", r.user, pr.GetNumber())
			if _, _, err := r.gh.Client.PullRequests.RequestReviewers(
				r.gh.Ctx,
				r.gh.Owner,
				r.gh.Repo,
				pr.GetNumber(),
				github.ReviewersRequest{
					Reviewers: []string{r.user},
				},
			); err != nil {
				r.gh.Logger.Errorf("Unable to request review from user %s on PR %d: %v", r.user, pr.GetNumber(), err)
			}
		}
	}

	// Check if user already approved this PR.
	reviews, err := r.gh.ListPRReviews(pr.GetNumber())
	if err != nil {
		r.gh.Logger.Errorf("unable to check if user %s already approved this PR: %v", r.user, err)
		return utils.AddStatusNode(false, detail, details)
	}

	for _, review := range reviews {
		if review.GetUser().GetLogin() == r.user {
			r.gh.Logger.Debugf("User %s already reviewed PR %d with state %s", r.user, pr.GetNumber(), review.GetState())
			return utils.AddStatusNode(review.GetState() == "APPROVED", detail, details)
		}
	}
	r.gh.Logger.Debugf("User %s has not reviewed PR %d yet", r.user, pr.GetNumber())

	return utils.AddStatusNode(false, detail, details)
}

func ReviewByUser(gh *client.GitHub, user string) Requirement {
	return &reviewByUser{gh, user}
}

// Reviewer Requirement.
type reviewByTeamMembers struct {
	gh    *client.GitHub
	team  string
	count uint
}

var _ Requirement = &reviewByTeamMembers{}

func (r *reviewByTeamMembers) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	detail := fmt.Sprintf("At least %d user(s) of the team %s approved pull request", r.count, r.team)

	// If not a dry run, make the user a reviewer if he's not already.
	if !r.gh.DryRun {
		requested := false
		reviewers, err := r.gh.ListPRReviewers(pr.GetNumber())
		if err != nil {
			r.gh.Logger.Errorf("unable to check if team %s review is already requested: %v", r.team, err)
			return utils.AddStatusNode(false, detail, details)
		}

		for _, team := range reviewers.Teams {
			if team.GetSlug() == r.team {
				requested = true
				break
			}
		}

		if requested {
			r.gh.Logger.Debugf("Review of team %s already requested on PR %d", r.team, pr.GetNumber())
		} else {
			r.gh.Logger.Debugf("Requesting review from team %s on PR %d", r.team, pr.GetNumber())
			if _, _, err := r.gh.Client.PullRequests.RequestReviewers(
				r.gh.Ctx,
				r.gh.Owner,
				r.gh.Repo,
				pr.GetNumber(),
				github.ReviewersRequest{
					TeamReviewers: []string{r.team},
				},
			); err != nil {
				r.gh.Logger.Errorf("Unable to request review from team %s on PR %d: %v", r.team, pr.GetNumber(), err)
			}
		}
	}

	// Check how many members of this team already approved this PR.
	approved := uint(0)
	reviews, err := r.gh.ListPRReviews(pr.GetNumber())
	if err != nil {
		r.gh.Logger.Errorf("unable to check if a member of team %s already approved this PR: %v", r.team, err)
		return utils.AddStatusNode(false, detail, details)
	}

	for _, review := range reviews {
		teamMembers, err := r.gh.ListTeamMembers(r.team)
		if err != nil {
			r.gh.Logger.Errorf(err.Error())
			continue
		}

		for _, member := range teamMembers {
			if review.GetUser().GetLogin() == member.GetLogin() {
				if review.GetState() == "APPROVED" {
					approved += 1
				}
				r.gh.Logger.Debugf("Member %s from team %s already reviewed PR %d with state %s (%d/%d required approval(s))", member.GetLogin(), r.team, pr.GetNumber(), review.GetState(), approved, r.count)
			}
		}
	}

	return utils.AddStatusNode(approved >= r.count, detail, details)
}

func ReviewByTeamMembers(gh *client.GitHub, team string, count uint) Requirement {
	return &reviewByTeamMembers{gh, team, count}
}
