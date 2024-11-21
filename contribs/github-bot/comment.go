package main

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/gnolang/gno/contribs/github-bot/client"

	"github.com/google/go-github/v64/github"
	"github.com/sethvargo/go-githubactions"
)

var errTriggeredByBot = errors.New("event triggered by bot")

// Compile regex only once
var (
	// Regex for capturing the entire line of a manual check
	manualCheckLine = regexp.MustCompile(`(?m:^- \[([ x])\] (.+)?$)`)
	// Regex for capturing only the user who checked it
	manualCheckDetails = regexp.MustCompile(`(?m:(.+) \(checked by @(\w+)\)$)`)
	// Regex for capturing only the checkboxes
	checkboxes = regexp.MustCompile(`(?m:^- \[[ x]\])`)
	// Regex used to capture markdown links
	markdownLink = regexp.MustCompile(`\[(.*)\]\(.*\)`)
)

// These structures contain the necessary information to generate
// the bot's comment from the template file
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
	AutoRules    []AutoContent
	ManualRules  []ManualContent
	allSatisfied bool
}

// getCommentManualChecks parses the bot comment to get the checkbox status,
// the check description and the username who checked it
func getCommentManualChecks(commentBody string) map[string][2]string {
	checks := make(map[string][2]string)

	// For each line that matches the "Manual check" regex
	for _, match := range manualCheckLine.FindAllStringSubmatch(commentBody, -1) {
		status := match[1]
		// Try to capture an occurence of : (checked by @user)
		if details := manualCheckDetails.FindAllStringSubmatch(match[2], -1); len(details) > 0 {
			// If found, set both the status and the user that checked the box
			description := details[0][1]
			checkedBy := details[0][2]
			checks[description] = [2]string{status, checkedBy}
			continue
		}

		// If not found, set only the status of the box
		description := match[2]
		checks[description] = [2]string{status}
	}

	return checks
}

// handleCommentUpdate checks if:
//   - the current run was triggered by GitHub Actions
//   - the triggering event is an edit of the bot comment
//   - the comment was not edited by the bot itself (prevent infinite loop)
//   - the comment change is only a checkbox being checked or unckecked (or restore it)
//   - the actor / comment editor has permission to modify this checkbox (or restore it)
func handleCommentUpdate(gh *client.GitHub) error {
	// Get GitHub Actions context to retrieve comment update
	actionCtx, err := githubactions.Context()
	if err != nil {
		gh.Logger.Debugf("Unable to retrieve GitHub Actions context: %v", err)
		return nil
	}

	// Ignore if it's not a comment related event
	if actionCtx.EventName != "issue_comment" {
		gh.Logger.Debugf("Event is not issue comment related (%s)", actionCtx.EventName)
		return nil
	}

	// Ignore if the action type is not deleted or edited
	actionType, ok := actionCtx.Event["action"].(string)
	if !ok {
		return errors.New("unable to get type on issue comment event")
	}

	if actionType != "deleted" && actionType != "edited" {
		return nil
	}

	// Return if comment was edited by bot (current authenticated user)
	authUser, _, err := gh.Client.Users.Get(gh.Ctx, "")
	if err != nil {
		return fmt.Errorf("unable to get authenticated user: %w", err)
	}

	if actionCtx.Actor == authUser.GetLogin() {
		gh.Logger.Debugf("Prevent infinite loop if the bot comment was edited by the bot itself")
		return errTriggeredByBot
	}

	// Ignore if comment edition author is not the bot
	comment, ok := actionCtx.Event["comment"].(map[string]any)
	if !ok {
		return errors.New("unable to get comment on issue comment event")
	}

	author, ok := comment["user"].(map[string]any)
	if !ok {
		return errors.New("unable to get comment user on issue comment event")
	}

	login, ok := author["login"].(string)
	if !ok {
		return errors.New("unable to get comment user login on issue comment event")
	}

	// If comment edition author is not the bot, return
	if login != authUser.GetLogin() {
		return nil
	}

	// Get comment current body
	current, ok := comment["body"].(string)
	if !ok {
		return errors.New("unable to get comment body on issue comment event")
	}

	// Get comment updated body
	changes, ok := actionCtx.Event["changes"].(map[string]any)
	if !ok {
		return errors.New("unable to get changes on issue comment event")
	}

	changesBody, ok := changes["body"].(map[string]any)
	if !ok {
		return errors.New("unable to get changes body on issue comment event")
	}

	previous, ok := changesBody["from"].(string)
	if !ok {
		return errors.New("unable to get changes body content on issue comment event")
	}

	// Get PR number from GitHub Actions context
	issue, ok := actionCtx.Event["issue"].(map[string]any)
	if !ok {
		return errors.New("unable to get issue on issue comment event")
	}

	num, ok := issue["number"].(float64)
	if !ok || num <= 0 {
		return errors.New("unable to get issue number on issue comment event")
	}

	// Check if change is only a checkbox being checked or unckecked
	if checkboxes.ReplaceAllString(current, "") != checkboxes.ReplaceAllString(previous, "") {
		// If not, restore previous comment body
		if !gh.DryRun {
			gh.SetBotComment(previous, int(num))
		}
		return errors.New("bot comment edited outside of checkboxes")
	}

	// Check if actor / comment editor has permission to modify changed boxes
	currentChecks := getCommentManualChecks(current)
	previousChecks := getCommentManualChecks(previous)
	edited := ""
	for key := range currentChecks {
		// If there is no diff for this check, ignore it
		if currentChecks[key][0] == previousChecks[key][0] {
			continue
		}

		// Get teams allowed to edit this box from config
		var teams []string
		found := false
		_, manualRules := config(gh)

		for _, manualRule := range manualRules {
			if manualRule.description == key {
				found = true
				teams = manualRule.teams
			}
		}

		// If rule were not found, return to reprocess the bot comment entirely
		// (maybe bot config was updated since last run?)
		if !found {
			gh.Logger.Debugf("Updated rule not found in config: %s", key)
			return nil
		}

		// If teams specified in rule, check if actor is a member of one of them
		if len(teams) > 0 {
			if gh.IsUserInTeams(actionCtx.Actor, teams) {
				if !gh.DryRun {
					gh.SetBotComment(previous, int(num))
				}
				return errors.New("checkbox edited by a user not allowed to")
			}
		}

		// This regex capture only the line of the current check
		specificManualCheck := regexp.MustCompile(fmt.Sprintf(`(?m:^- \[%s\] %s.*$)`, currentChecks[key][0], key))

		// If the box is checked, append the username of the user who checked it
		if strings.TrimSpace(currentChecks[key][0]) == "x" {
			replacement := fmt.Sprintf("- [%s] %s (checked by @%s)", currentChecks[key][0], key, actionCtx.Actor)
			edited = specificManualCheck.ReplaceAllString(current, replacement)
		} else { // Else, remove the username of the user
			replacement := fmt.Sprintf("- [%s] %s", currentChecks[key][0], key)
			edited = specificManualCheck.ReplaceAllString(current, replacement)
		}
	}

	// Update comment with username
	if edited != "" && !gh.DryRun {
		gh.SetBotComment(edited, int(num))
		gh.Logger.Debugf("Comment manual checks updated successfully")
	}

	return nil
}

// generateComment generates a comment using the template file and the
// content passed as parameter
func generateComment(content CommentContent) (string, error) {
	// Custom function to strip markdown links
	funcMap := template.FuncMap{
		"stripLinks": func(input string) string {
			return markdownLink.ReplaceAllString(input, "$1")
		},
	}

	// Bind markdown stripping function to template generator
	const tmplFile = "comment.tmpl"
	tmpl, err := template.New(tmplFile).Funcs(funcMap).ParseFiles(tmplFile)
	if err != nil {
		return "", fmt.Errorf("unable to init template: %v", err)
	}

	// Generate bot comment using template file
	var commentBytes bytes.Buffer
	if err := tmpl.Execute(&commentBytes, content); err != nil {
		return "", fmt.Errorf("unable to execute template: %v", err)
	}

	return commentBytes.String(), nil
}

// updatePullRequest updates or creates both the bot comment and the commit status
func updatePullRequest(gh *client.GitHub, pr *github.PullRequest, content CommentContent) error {
	// Generate comment text content
	commentText, err := generateComment(content)
	if err != nil {
		return fmt.Errorf("unable to generate comment on PR %d: %v", pr.GetNumber(), err)
	}

	// Update comment on pull request
	comment := gh.SetBotComment(commentText, pr.GetNumber())
	if comment != nil {
		gh.Logger.Infof("Comment successfully updated on PR %d", pr.GetNumber())
	}

	// Prepare commit status content
	var (
		context     = "Merge Requirements"
		targetURL   = ""
		state       = "pending"
		description = "Some requirements are not satisfied yet. See bot comment."
	)

	if comment != nil {
		targetURL = comment.GetHTMLURL()
	}

	if content.allSatisfied {
		state = "success"
		description = "All requirements are satisfied."
	}

	// Update or create commit status
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
		return fmt.Errorf("unable to create status on PR %d: %v", pr.GetNumber(), err)
	} else {
		gh.Logger.Infof("Commit status successfully updated on PR %d", pr.GetNumber())
	}

	return nil
}
