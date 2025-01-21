package requirements

import (
	"fmt"
	"slices"

	"github.com/gnolang/gno/contribs/github-bot/internal/client"
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

// Reviewer Requirement.
type approvalByUser struct {
	gh   *client.GitHub
	user string
}

var _ Requirement = &approvalByUser{}

func (r *approvalByUser) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
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
			return utils.AddStatusNode(review.GetState() == utils.ReviewStateApproved, detail, details)
		}
	}
	r.gh.Logger.Debugf("User %s has not reviewed PR %d yet", r.user, pr.GetNumber())

	return utils.AddStatusNode(false, detail, details)
}

func ApprovalByUser(gh *client.GitHub, user string) Requirement {
	return &approvalByUser{gh, user}
}

// Reviewer Requirement.
type ReviewByTeamMembersRequirement struct {
	gh           *client.GitHub
	team         string
	count        uint
	desiredState string
}

var _ Requirement = &ReviewByTeamMembersRequirement{}

func (r *ReviewByTeamMembersRequirement) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	detail := fmt.Sprintf("At least %d user(s) of the team %s reviewed pull request", r.count, r.team)
	if r.desiredState != "" {
		detail += fmt.Sprintf("(with state %q)", r.desiredState)
	}

	teamMembers, err := r.gh.ListTeamMembers(r.team)
	if err != nil {
		r.gh.Logger.Errorf(err.Error())
		return utils.AddStatusNode(false, detail, details)
	}

	// If not a dry run, request a team review if no member has reviewed yet,
	// and the team review has not been requested.
	if !r.gh.DryRun {
		var teamRequested bool
		var usersRequested []string

		reviewers, err := r.gh.ListPRReviewers(pr.GetNumber())
		if err != nil {
			r.gh.Logger.Errorf("unable to check if team %s review is already requested: %v", r.team, err)
			return utils.AddStatusNode(false, detail, details)
		}

		for _, team := range reviewers.Teams {
			if team.GetSlug() == r.team {
				teamRequested = true
				break
			}
		}

		if !teamRequested {
			for _, user := range reviewers.Users {
				if slices.ContainsFunc(teamMembers, func(memb *github.User) bool {
					return memb.GetID() == user.GetID()
				}) {
					usersRequested = append(usersRequested, user.GetLogin())
				}
			}
		}

		switch {
		case teamRequested:
			r.gh.Logger.Debugf("Review of team %s already requested on PR %d", r.team, pr.GetNumber())
		case len(usersRequested) > 0:
			r.gh.Logger.Debugf("Members %v of team %s already requested on (or reviewed) PR %d",
				usersRequested, r.team, pr.GetNumber())
		default:
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

	// Check how many members of this team already reviewed this PR.
	reviewCount := uint(0)
	reviews, err := r.gh.ListPRReviews(pr.GetNumber())
	if err != nil {
		r.gh.Logger.Errorf("unable to check if a member of team %s already reviewed this PR: %v", r.team, err)
		return utils.AddStatusNode(false, detail, details)
	}

	stateStr := ""
	if r.desiredState != "" {
		stateStr = fmt.Sprintf("%q ", r.desiredState)
	}
	for _, review := range reviews {
		for _, member := range teamMembers {
			if review.GetUser().GetLogin() == member.GetLogin() {
				if desired := r.desiredState; desired == "" || desired == review.GetState() {
					reviewCount += 1
				}
				r.gh.Logger.Debugf(
					"Member %s from team %s already reviewed PR %d with state %s (%d/%d required %sreview(s))",
					member.GetLogin(), r.team, pr.GetNumber(), review.GetState(), reviewCount, r.count, stateStr,
				)
			}
		}
	}

	return utils.AddStatusNode(reviewCount >= r.count, detail, details)
}

// WithCount specifies the number of required reviews.
// By default, this is 1.
func (r *ReviewByTeamMembersRequirement) WithCount(n uint) *ReviewByTeamMembersRequirement {
	if n < 1 {
		panic("number of required reviews should be at least 1")
	}
	r.count = n
	return r
}

// WithDesiredState specifies the desired state of the PR reviews.
//
// If an empty string is passed, then all reviews are counted. This is the default.
func (r *ReviewByTeamMembersRequirement) WithDesiredState(state string) *ReviewByTeamMembersRequirement {
	r.desiredState = state
	return r
}

// ReviewByTeamMembers specifies that the given pull request should receive at
// least one review from a member of the given team.
//
// The number of required reviews, or the state of the reviews (e.g., to filter
// only for approval reviews) can be modified using WithCount and WithDesiredState.
func ReviewByTeamMembers(gh *client.GitHub, team string) *ReviewByTeamMembersRequirement {
	return &ReviewByTeamMembersRequirement{
		gh:    gh,
		team:  team,
		count: 1,
	}
}

type approvalByOrgMembers struct {
	gh    *client.GitHub
	count uint
}

var _ Requirement = &approvalByOrgMembers{}

func (r *approvalByOrgMembers) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	detail := fmt.Sprintf("At least %d user(s) of the organization approved the pull request", r.count)

	// Check how many members of this team already approved this PR.
	approved := uint(0)
	reviews, err := r.gh.ListPRReviews(pr.GetNumber())
	if err != nil {
		r.gh.Logger.Errorf("unable to check number of reviews on this PR: %v", err)
		return utils.AddStatusNode(false, detail, details)
	}

	for _, review := range reviews {
		if review.GetAuthorAssociation() == "MEMBER" {
			if review.GetState() == utils.ReviewStateApproved {
				approved++
			}
			r.gh.Logger.Debugf(
				"Member %s already reviewed PR %d with state %s (%d/%d required approval(s))",
				review.GetUser().GetLogin(), pr.GetNumber(), review.GetState(),
				approved, r.count,
			)
		}
	}

	return utils.AddStatusNode(approved >= r.count, detail, details)
}

func ApprovalByOrgMembers(gh *client.GitHub, count uint) Requirement {
	return &approvalByOrgMembers{gh, count}
}
