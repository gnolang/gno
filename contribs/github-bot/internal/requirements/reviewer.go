package requirements

import (
	"fmt"
	"slices"

	"github.com/gnolang/gno/contribs/github-bot/internal/client"
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

// RequestAction controls what to do about the review request.
type RequestAction byte

const (
	// RequestApply will request a review from the user/team if not already requested.
	RequestApply = iota
	// RequestRemove will remove the review request from the user/team.
	RequestRemove
	// RequestIgnore always leaves the review request as it is.
	RequestIgnore
)

// deduplicateReviews returns a list of reviews with at most 1 review per
// author, where approval/changes requested reviews are preferred over comments
// and later reviews are preferred over earlier ones.
func deduplicateReviews(reviews []*github.PullRequestReview) []*github.PullRequestReview {
	added := make(map[string]int)
	result := make([]*github.PullRequestReview, 0, len(reviews))
	for _, rev := range reviews {
		idx, ok := added[rev.User.GetLogin()]
		switch utils.ReviewState(rev.GetState()) {
		case utils.ReviewStateApproved, utils.ReviewStateChangesRequested:
			// this review changes the "approval state", and is more relevant,
			// so substitute it with the previous one if it exists.
			if ok {
				result[idx] = rev
			} else {
				result = append(result, rev)
				added[rev.User.GetLogin()] = len(result) - 1
			}
		case utils.ReviewStateCommented:
			// this review does not change the "approval state", so only append
			// it if a previous review doesn't exist.
			if !ok {
				result = append(result, rev)
				added[rev.User.GetLogin()] = len(result) - 1
			}
		case utils.ReviewStateDismissed:
			// this state just dismisses any previous review, so remove previous
			// entry for this user if it exists.
			if ok {
				result[idx] = nil
			}
		default:
			panic(fmt.Sprintf("invalid review state %q", rev.GetState()))
		}
	}
	// Remove nil entries from the result (dismissed reviews).
	result = slices.DeleteFunc(result, func(r *github.PullRequestReview) bool {
		return r == nil
	})

	return result
}

// ReviewByUserRequirement asserts that there is a review by the given user,
// and if given that the review matches the desiredState.
type ReviewByUserRequirement struct {
	gh           *client.GitHub
	user         string
	desiredState string
	action       RequestAction
}

var _ Requirement = &ReviewByUserRequirement{}

// IsSatisfied implements [Requirement].
func (r *ReviewByUserRequirement) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	detail := fmt.Sprintf("This user reviewed pull request: %s", r.user)
	if r.desiredState != "" {
		detail += fmt.Sprintf(" (with state %q)", r.desiredState)
	}

	// Check if user already approved this PR.
	reviews, err := r.gh.ListPRReviews(pr.GetNumber())
	if err != nil {
		r.gh.Logger.Errorf("unable to check if user %s already approved this PR: %v", r.user, err)
		return utils.AddStatusNode(false, detail, details)
	}
	reviews = deduplicateReviews(reviews)

	for _, review := range reviews {
		if review.GetUser().GetLogin() == r.user {
			r.gh.Logger.Debugf("User %s already reviewed PR %d with state %s", r.user, pr.GetNumber(), review.GetState())
			result := r.desiredState == "" || review.GetState() == r.desiredState
			return utils.AddStatusNode(result, detail, details)
		}
	}
	r.gh.Logger.Debugf("User %s has not reviewed PR %d yet", r.user, pr.GetNumber())

	// If not a dry run, change the review request according to the action.
	if !r.gh.DryRun && r.action != RequestIgnore {
		// Check if this user is already requested for review.
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

		switch r.action {
		case RequestApply:
			switch {
			case requested:
				r.gh.Logger.Debugf("Review of user %s already requested on PR %d", r.user, pr.GetNumber())
			case r.user == pr.GetUser().GetLogin():
				r.gh.Logger.Debugf("Review of user %s is not requested on PR %d because he's the author", r.user, pr.GetNumber())
			default:
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
		case RequestRemove:
			switch {
			case !requested:
				r.gh.Logger.Debugf("Review of user %s already not requested on PR %d", r.user, pr.GetNumber())
			default:
				r.gh.Logger.Debugf("Removing review request from user %s on PR %d", r.user, pr.GetNumber())
				if _, err := r.gh.Client.PullRequests.RemoveReviewers(
					r.gh.Ctx,
					r.gh.Owner,
					r.gh.Repo,
					pr.GetNumber(),
					github.ReviewersRequest{
						Reviewers: []string{r.user},
					},
				); err != nil {
					r.gh.Logger.Errorf("Unable to remove review request from user %s on PR %d: %v", r.user, pr.GetNumber(), err)
				}
			}
		}
	}

	return utils.AddStatusNode(false, detail, details)
}

// WithDesiredState asserts that the review by the given user should also be
// of the given ReviewState.
//
// If an empty string is passed, then all reviews are counted. This is the default.
func (r *ReviewByUserRequirement) WithDesiredState(s utils.ReviewState) *ReviewByUserRequirement {
	if s != "" && !s.Valid() {
		panic(fmt.Sprintf("invalid state: %q", s))
	}
	r.desiredState = string(s)
	return r
}

// ReviewByUser asserts that the PR has been reviewed by the given user.
func ReviewByUser(gh *client.GitHub, user string, action RequestAction) *ReviewByUserRequirement {
	return &ReviewByUserRequirement{
		gh:     gh,
		user:   user,
		action: action,
	}
}

// ReviewByTeamMembersRequirement asserts that count members of the given team
// have reviewed the PR. Additionally, using desiredState, it may be required
// that the PR reviews be of that state.
type ReviewByTeamMembersRequirement struct {
	gh           *client.GitHub
	team         string
	count        uint
	desiredState string
	action       RequestAction
}

var _ Requirement = &ReviewByTeamMembersRequirement{}

// IsSatisfied implements [Requirement].
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

	reviews, err := r.gh.ListPRReviews(pr.GetNumber())
	if err != nil {
		r.gh.Logger.Errorf("unable to fetch existing reviews of pr %d: %v", pr.GetNumber(), err)
		return utils.AddStatusNode(false, detail, details)
	}

	reviews = deduplicateReviews(reviews)

	// If not a dry run, request a team review if no member has reviewed yet,
	// and the team review has not been requested.
	if !r.gh.DryRun && r.action != RequestIgnore {
		// Check if the team or any of its members are already requested for review.
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
				if containsUserWithLogin(teamMembers, user.GetLogin()) {
					usersRequested = append(usersRequested, user.GetLogin())
				}
			}

			for _, rev := range reviews {
				// if not already requested and user is a team member...
				if !slices.Contains(usersRequested, rev.User.GetLogin()) &&
					containsUserWithLogin(teamMembers, rev.User.GetLogin()) {
					usersRequested = append(usersRequested, rev.User.GetLogin())
				}
			}
		}

		switch r.action {
		case RequestApply:
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
		case RequestRemove:
			switch {
			case !teamRequested:
				r.gh.Logger.Debugf("Review of team %s already not requested on PR %d", r.team, pr.GetNumber())
			default:
				r.gh.Logger.Debugf("Removing review request from team %s on PR %d", r.team, pr.GetNumber())
				if _, err := r.gh.Client.PullRequests.RemoveReviewers(
					r.gh.Ctx,
					r.gh.Owner,
					r.gh.Repo,
					pr.GetNumber(),
					github.ReviewersRequest{
						TeamReviewers: []string{r.team},
					},
				); err != nil {
					r.gh.Logger.Errorf("Unable to remove review request from team %s on PR %d: %v", r.team, pr.GetNumber(), err)
				}
			}
		}
	}

	// Check how many members of this team already reviewed this PR.
	reviewCount := uint(0)

	for _, review := range reviews {
		login := review.GetUser().GetLogin()
		if containsUserWithLogin(teamMembers, login) {
			if desired := r.desiredState; desired == "" || desired == review.GetState() {
				reviewCount += 1
			}
			r.gh.Logger.Debugf(
				"Member %s from team %s already reviewed PR %d with state %s (%d/%d required review(s) with state %q)",
				login, r.team, pr.GetNumber(), review.GetState(), reviewCount, r.count, r.desiredState,
			)
		}
	}

	return utils.AddStatusNode(reviewCount >= r.count, detail, details)
}

func containsUserWithLogin(users []*github.User, login string) bool {
	return slices.ContainsFunc(users, func(u *github.User) bool {
		return u.GetLogin() == login
	})
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

// WithDesiredState asserts that the reviews should also be of the given ReviewState.
//
// If an empty string is passed, then all reviews are counted. This is the default.
func (r *ReviewByTeamMembersRequirement) WithDesiredState(s utils.ReviewState) *ReviewByTeamMembersRequirement {
	if s != "" && !s.Valid() {
		panic(fmt.Sprintf("invalid state: %q", s))
	}
	r.desiredState = string(s)
	return r
}

// ReviewByTeamMembers specifies that the given pull request should receive at
// least one review from a member of the given team.
//
// The number of required reviews, or the state of the reviews (e.g., to filter
// only for approval reviews) can be modified using WithCount and WithDesiredState.
func ReviewByTeamMembers(gh *client.GitHub, team string, action RequestAction) *ReviewByTeamMembersRequirement {
	return &ReviewByTeamMembersRequirement{
		gh:     gh,
		team:   team,
		count:  1,
		action: action,
	}
}

// ReviewByOrgMembersRequirement asserts that the given PR has been reviewed by
// at least count members of the given organization, filtering for PR reviews
// with state desiredState.
type ReviewByOrgMembersRequirement struct {
	gh           *client.GitHub
	count        uint
	desiredState string
}

var _ Requirement = &ReviewByOrgMembersRequirement{}

// IsSatisfied implements [Requirement].
func (r *ReviewByOrgMembersRequirement) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	detail := fmt.Sprintf("At least %d user(s) of the organization reviewed the pull request", r.count)
	if r.desiredState != "" {
		detail += fmt.Sprintf(" (with state %q)", r.desiredState)
	}

	// Check how many members of this team already reviewed this PR.
	reviewed := uint(0)
	reviews, err := r.gh.ListPRReviews(pr.GetNumber())
	if err != nil {
		r.gh.Logger.Errorf("unable to check number of reviews on this PR: %v", err)
		return utils.AddStatusNode(false, detail, details)
	}
	reviews = deduplicateReviews(reviews)

	for _, review := range reviews {
		if review.GetAuthorAssociation() == "MEMBER" {
			if r.desiredState == "" || review.GetState() == r.desiredState {
				reviewed++
			}
			r.gh.Logger.Debugf(
				"Member %s already reviewed PR %d with state %s (%d/%d required reviews with state %q)",
				review.GetUser().GetLogin(), pr.GetNumber(), review.GetState(),
				reviewed, r.count, r.desiredState,
			)
		}
	}

	return utils.AddStatusNode(reviewed >= r.count, detail, details)
}

// WithCount specifies the number of required reviews.
// By default, this is 1.
func (r *ReviewByOrgMembersRequirement) WithCount(n uint) *ReviewByOrgMembersRequirement {
	if n < 1 {
		panic("number of required reviews should be at least 1")
	}
	r.count = n
	return r
}

// WithDesiredState asserts that the reviews should also be of the given ReviewState.
//
// If an empty string is passed, then all reviews are counted. This is the default.
func (r *ReviewByOrgMembersRequirement) WithDesiredState(s utils.ReviewState) *ReviewByOrgMembersRequirement {
	if s != "" && !s.Valid() {
		panic(fmt.Sprintf("invalid state: %q", s))
	}
	r.desiredState = string(s)
	return r
}

// ReviewByOrgMembers asserts that at least 1 member of the organization
// reviewed this PR.
func ReviewByOrgMembers(gh *client.GitHub) *ReviewByOrgMembersRequirement {
	return &ReviewByOrgMembersRequirement{gh: gh, count: 1}
}

// ReviewByAnyUserRequirement asserts that the given PR has been reviewed by
// at least one of the given users, filtering for PR reviews with state desiredState.
type ReviewByAnyUserRequirement struct {
	gh           *client.GitHub
	users        []string
	desiredState string
}

var _ Requirement = &ReviewByAnyUserRequirement{}

// IsSatisfied implements [Requirement].
func (r *ReviewByAnyUserRequirement) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	detail := fmt.Sprintf("At least one of these user(s) reviewed the pull request: %v", r.users)
	if r.desiredState != "" {
		detail += fmt.Sprintf(" (with state %q)", r.desiredState)
	}

	// Check if one of the users already approved this PR.
	reviews, err := r.gh.ListPRReviews(pr.GetNumber())
	if err != nil {
		r.gh.Logger.Errorf("unable to check number of reviews on this PR: %v", err)
		return utils.AddStatusNode(false, detail, details)
	}
	reviews = deduplicateReviews(reviews)

	for _, review := range reviews {
		if r.desiredState == "" || review.GetState() == r.desiredState {
			for _, user := range r.users {
				if review.GetUser().GetLogin() == user {
					detail = fmt.Sprintf("User %s already reviewed PR %d with state %s", user, pr.GetNumber(), review.GetState())
					r.gh.Logger.Debugf("%s", detail)
					return utils.AddStatusNode(true, detail, details)
				}
			}
		}
	}

	return utils.AddStatusNode(false, detail, details)
}

// WithDesiredState asserts that the matching review should also be of the given ReviewState.
//
// If an empty string is passed, then all reviews are counted. This is the default.
func (r *ReviewByAnyUserRequirement) WithDesiredState(s utils.ReviewState) *ReviewByAnyUserRequirement {
	if s != "" && !s.Valid() {
		panic(fmt.Sprintf("invalid state: %q", s))
	}
	r.desiredState = string(s)
	return r
}

// ReviewByAnyUser asserts that at least one of the given users reviewed this PR.
func ReviewByAnyUser(gh *client.GitHub, users ...string) *ReviewByAnyUserRequirement {
	return &ReviewByAnyUserRequirement{gh: gh, users: users}
}
