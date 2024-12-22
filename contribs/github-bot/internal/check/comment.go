package check

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/gnolang/gno/contribs/github-bot/internal/client"
	"github.com/gnolang/gno/contribs/github-bot/internal/config"
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"
	"github.com/google/go-github/v64/github"
	"github.com/sethvargo/go-githubactions"
)

//go:embed comment.tmpl
var tmplString string // Embed template used for comment generation.

var errTriggeredByBot = errors.New("event triggered by bot")

// Compile regex only once.
var (
	// Regex for capturing the entire line of a manual check.
	manualCheckLine = regexp.MustCompile(`(?m:^- \[([ xX])\] (.+?)(?: \(checked by @([A-Za-z0-9-]+)\))?$)`)
	// Regex for capturing only the checkboxes.
	checkboxes = regexp.MustCompile(`(?m:^- \[[ xX]\])`)
	// Regex used to capture markdown links.
	markdownLink = regexp.MustCompile(`\[(.*)\]\([^)]*\)`)
)

// These structures contain the necessary information to generate
// the bot's comment from the template file.
type AutoContent struct {
	Description        string
	Satisfied          bool
	ConditionDetails   string
	RequirementDetails string
}
type ManualContent struct {
	Description      string
	CheckedBy        string
	ConditionDetails string
	Teams            []string
}
type CommentContent struct {
	AutoRules          []AutoContent
	ManualRules        []ManualContent
	AutoAllSatisfied   bool
	ManualAllSatisfied bool
	ForceSkip          bool
}

type manualCheckDetails struct {
	status    string
	checkedBy string
}

// getCommentManualChecks parses the bot comment to get the checkbox status,
// the check description and the username who checked it.
func getCommentManualChecks(commentBody string) map[string]manualCheckDetails {
	checks := make(map[string]manualCheckDetails)

	// For each line that matches the "Manual check" regex.
	for _, match := range manualCheckLine.FindAllStringSubmatch(commentBody, -1) {
		description := match[2]
		status := strings.ToLower(match[1]) // if X captured, convert it to x.
		checkedBy := ""
		if len(match) > 3 {
			checkedBy = match[3]
		}

		checks[description] = manualCheckDetails{status: status, checkedBy: checkedBy}
	}

	return checks
}

// handleCommentUpdate checks if:
//   - the current run was triggered by GitHub Actions
//   - the triggering event is an edit of the bot comment
//   - the comment was not edited by the bot itself (prevent infinite loop)
//   - the comment change is only a checkbox being checked or unckecked (or restore it)
//   - the actor / comment editor has permission to modify this checkbox (or restore it)
func handleCommentUpdate(gh *client.GitHub, actionCtx *githubactions.GitHubContext) error {
	// Ignore if it's not a comment related event.
	if actionCtx.EventName != utils.EventIssueComment {
		gh.Logger.Debugf("Event is not issue comment related (%s)", actionCtx.EventName)
		return nil
	}

	// Ignore if the action type is not deleted or edited.
	actionType, ok := actionCtx.Event["action"].(string)
	if !ok {
		return errors.New("unable to get type on issue comment event")
	}

	if actionType != "deleted" && actionType != "edited" {
		return nil
	}

	// Get PR number from GitHub Actions context.
	prNumFloat, ok := utils.IndexMap(actionCtx.Event, "issue", "number").(float64)
	if !ok || prNumFloat <= 0 {
		return errors.New("unable to get issue number on issue comment event")
	}
	prNum := int(prNumFloat)

	// Ignore if this comment update is not related to an opened PR.
	if _, err := gh.GetOpenedPullRequest(prNum); err != nil {
		return nil // May come from an issue or a closed PR
	}

	// Return if comment was edited by bot (current authenticated user).
	authUser, _, err := gh.Client.Users.Get(gh.Ctx, "")
	if err != nil {
		return fmt.Errorf("unable to get authenticated user: %w", err)
	}

	if actionCtx.Actor == authUser.GetLogin() {
		gh.Logger.Debugf("Prevent infinite loop if the bot comment was edited by the bot itself")
		return errTriggeredByBot
	}

	// Get login of the author of the edited comment.
	login, ok := utils.IndexMap(actionCtx.Event, "comment", "user", "login").(string)
	if !ok {
		return errors.New("unable to get comment user login on issue comment event")
	}

	// If the author is not the bot, return.
	if login != authUser.GetLogin() {
		return nil
	}

	// Get comment updated body.
	current, ok := utils.IndexMap(actionCtx.Event, "comment", "body").(string)
	if !ok {
		return errors.New("unable to get comment body on issue comment event")
	}

	// Get comment previous body.
	previous, ok := utils.IndexMap(actionCtx.Event, "changes", "body", "from").(string)
	if !ok {
		return errors.New("unable to get changes body content on issue comment event")
	}

	// Check if change is only a checkbox being checked or unckecked.
	if checkboxes.ReplaceAllString(current, "") != checkboxes.ReplaceAllString(previous, "") {
		// If not, restore previous comment body.
		if !gh.DryRun {
			gh.SetBotComment(previous, prNum)
		}
		return errors.New("bot comment edited outside of checkboxes")
	}

	// Check if actor / comment editor has permission to modify changed boxes.
	currentChecks := getCommentManualChecks(current)
	previousChecks := getCommentManualChecks(previous)
	edited := ""
	for key := range currentChecks {
		// If there is no diff for this check, ignore it.
		if currentChecks[key].status == previousChecks[key].status {
			continue
		}

		// Get teams allowed to edit this box from config.
		var teams []string
		found := false
		_, manualRules := config.Config(gh)

		for _, manualRule := range manualRules {
			if manualRule.Description == key {
				found = true
				teams = manualRule.Teams
			}
		}

		// If rule were not found, return to reprocess the bot comment entirely
		// (maybe bot config was updated since last run?).
		if !found {
			gh.Logger.Debugf("Updated rule not found in config: %s", key)
			return nil
		}

		// If teams specified in rule, check if actor is a member of one of them.
		if len(teams) > 0 {
			if !gh.IsUserInTeams(actionCtx.Actor, teams) { // If user not allowed to check the boxes.
				if !gh.DryRun {
					gh.SetBotComment(previous, prNum) // Then restore previous state.
				}
				return errors.New("checkbox edited by a user not allowed to")
			}
		}

		// This regex capture only the line of the current check.
		specificManualCheck := regexp.MustCompile(fmt.Sprintf(`(?m:^- \[%s\] %s.*$)`, currentChecks[key].status, regexp.QuoteMeta(key)))

		// If the box is checked, append the username of the user who checked it.
		if strings.TrimSpace(currentChecks[key].status) == "x" {
			replacement := fmt.Sprintf("- [%s] %s (checked by @%s)", currentChecks[key].status, key, actionCtx.Actor)
			edited = specificManualCheck.ReplaceAllString(current, replacement)
		} else {
			// Else, remove the username of the user.
			replacement := fmt.Sprintf("- [%s] %s", currentChecks[key].status, key)
			edited = specificManualCheck.ReplaceAllString(current, replacement)
		}
	}

	// Update comment with username.
	if edited != "" && !gh.DryRun {
		gh.SetBotComment(edited, prNum)
		gh.Logger.Debugf("Comment manual checks updated successfully")
	}

	return nil
}

// generateComment generates a comment using the template file and the
// content passed as parameter.
func generateComment(content CommentContent) (string, error) {
	// Custom function to strip markdown links.
	funcMap := template.FuncMap{
		"stripLinks": func(input string) string {
			return markdownLink.ReplaceAllString(input, "$1")
		},
	}

	// Bind markdown stripping function to template generator.
	tmpl, err := template.New("comment").Funcs(funcMap).Parse(tmplString)
	if err != nil {
		return "", fmt.Errorf("unable to init template: %w", err)
	}

	// Generate bot comment using template file.
	var commentBytes bytes.Buffer
	if err := tmpl.Execute(&commentBytes, content); err != nil {
		return "", fmt.Errorf("unable to execute template: %w", err)
	}

	return commentBytes.String(), nil
}

// updatePullRequest updates or creates both the bot comment and the commit status.
func updatePullRequest(gh *client.GitHub, pr *github.PullRequest, content CommentContent) error {
	// Generate comment text content.
	commentText, err := generateComment(content)
	if err != nil {
		return fmt.Errorf("unable to generate comment on PR %d: %w", pr.GetNumber(), err)
	}

	// Update comment on pull request.
	comment, err := gh.SetBotComment(commentText, pr.GetNumber())
	if err != nil {
		return fmt.Errorf("unable to update comment on PR %d: %w", pr.GetNumber(), err)
	} else {
		gh.Logger.Infof("Comment successfully updated on PR %d", pr.GetNumber())
	}

	// Prepare commit status content.
	var (
		context     = "Merge Requirements"
		targetURL   = comment.GetHTMLURL()
		state       = "success"
		description = "All requirements are satisfied."
	)

	if content.ForceSkip {
		description = "Bot checks are skipped for this PR."
	} else if !content.AutoAllSatisfied || !content.ManualAllSatisfied {
		state = "failure"
		description = "Some requirements are not satisfied yet. See bot comment."
	}

	// Update or create commit status.
	if _, _, err := gh.Client.Repositories.CreateStatus(
		gh.Ctx,
		gh.Owner,
		gh.Repo,
		pr.GetHead().GetSHA(),
		&github.RepoStatus{
			Context:     &context,
			State:       &state,
			TargetURL:   &targetURL,
			Description: &description,
		}); err != nil {
		return fmt.Errorf("unable to create status on PR %d: %w", pr.GetNumber(), err)
	} else {
		gh.Logger.Infof("Commit status successfully updated on PR %d", pr.GetNumber())
	}

	return nil
}
