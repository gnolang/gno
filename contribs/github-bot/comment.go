package main

import (
	"bot/client"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/google/go-github/v66/github"
	"github.com/sethvargo/go-githubactions"
)

type AutoContent struct {
	Description        string
	Satisfied          bool
	ConditionDetails   string
	RequirementDetails string
}
type ManualContent struct {
	Description      string
	ConditionDetails string
	CheckedBy        string
	Teams            []string
}

type CommentContent struct {
	AutoRules   []AutoContent
	ManualRules []ManualContent
}

// getCommentManualChecks parses the bot comment to get both the checkbox status,
// the check description and the username who checked it
func getCommentManualChecks(commentBody string) map[string][2]string {
	checks := make(map[string][2]string)

	reg := regexp.MustCompile(`(?m:^- \[([ x])\] (.+)?$)`)
	subReg := regexp.MustCompile(`(?m:(.+) \(checked by @(\w+)\)$)`)
	matches := reg.FindAllStringSubmatch(commentBody, -1)

	for _, match := range matches {
		if subMatches := subReg.FindAllStringSubmatch(match[2], -1); len(subMatches) > 0 {
			checks[subMatches[0][1]] = [2]string{match[1], subMatches[0][2]}
		} else {
			checks[match[2]] = [2]string{match[1]}
		}
	}

	return checks
}

// handleCommentUpdate checks if :
//   - the current run was triggered by GitHub Actions
//   - the triggering event is an edit of the bot comment
//   - the comment was not edited by the bot itself (prevent infinite loop)
//   - the comment change is only a checkbox being checked or unckecked (or restore it)
//   - the actor / comment editor has permission to modify this checkbox (or restore it)
func handleCommentUpdate(gh *client.GitHub) {
	// Get GitHub Actions context to retrieve comment update
	actionCtx, err := githubactions.Context()
	if err != nil {
		gh.Logger.Debugf("Unable to retrieve GitHub Actions context : %v", err)
		return
	}

	// Ignore if it's not an comment related event
	if actionCtx.EventName != "issue_comment" {
		gh.Logger.Debugf("Event is not issue comment related : %s", actionCtx.EventName)
		return
	}

	// Ignore if action type is not deleted or edited
	actionType, ok := actionCtx.Event["action"].(string)
	if !ok {
		gh.Logger.Errorf("Unable to get type on issue comment event")
		os.Exit(1)
	}

	if actionType != "deleted" && actionType != "edited" {
		return
	}

	// Exit if comment was edited by bot (current authenticated user)
	authUser, _, err := gh.Client.Users.Get(gh.Ctx, "")
	if err != nil {
		gh.Logger.Errorf("Unable to get authenticated user : %v", err)
		os.Exit(1)
	}

	if actionCtx.Actor == authUser.GetLogin() {
		gh.Logger.Debugf("Prevent infinite loop if the bot comment was edited by the bot itself")
		os.Exit(0)
	}

	// Ignore if edited comment author is not the bot
	comment, ok := actionCtx.Event["comment"].(map[string]any)
	if !ok {
		gh.Logger.Errorf("Unable to get comment on issue comment event")
		os.Exit(1)
	}

	author, ok := comment["user"].(map[string]any)
	if !ok {
		gh.Logger.Errorf("Unable to get comment user on issue comment event")
		os.Exit(1)
	}

	login, ok := author["login"].(string)
	if !ok {
		gh.Logger.Errorf("Unable to get comment user login on issue comment event")
		os.Exit(1)
	}

	if login != authUser.GetLogin() {
		return
	}

	// Get comment current body
	current, ok := comment["body"].(string)
	if !ok {
		gh.Logger.Errorf("Unable to get comment body on issue comment event")
		os.Exit(1)
	}

	// Get comment updated body
	changes, ok := actionCtx.Event["changes"].(map[string]any)
	if !ok {
		gh.Logger.Errorf("Unable to get changes on issue comment event")
		os.Exit(1)
	}

	changesBody, ok := changes["body"].(map[string]any)
	if !ok {
		gh.Logger.Errorf("Unable to get changes body on issue comment event")
		os.Exit(1)
	}

	previous, ok := changesBody["from"].(string)
	if !ok {
		gh.Logger.Errorf("Unable to get changes body content on issue comment event")
		os.Exit(1)
	}

	// Get PR number from GitHub Actions context
	issue, ok := actionCtx.Event["issue"].(map[string]any)
	if !ok {
		gh.Logger.Errorf("Unable to get issue on issue comment event")
		os.Exit(1)
	}

	num, ok := issue["number"].(float64)
	if !ok || num <= 0 {
		gh.Logger.Errorf("Unable to get issue number on issue comment event")
		os.Exit(1)
	}

	// Check if change is only a checkbox being checked or unckecked
	checkboxes := regexp.MustCompile(`(?m:^- \[[ x]\])`)
	if checkboxes.ReplaceAllString(current, "") != checkboxes.ReplaceAllString(previous, "") {
		// If not, restore previous comment body
		gh.Logger.Errorf("Bot comment edited outside of checkboxes")
		gh.SetBotComment(previous, int(num))
		os.Exit(1)
	}

	// Check if actor / comment editor has permission to modify changed boxes
	currentChecks := getCommentManualChecks(current)
	previousChecks := getCommentManualChecks(previous)
	edited := ""
	for key := range currentChecks {
		if currentChecks[key][0] != previousChecks[key][0] {
			// Get teams allowed to edit this box from config
			var teams []string
			found := false
			_, manualRules := config(gh)

			for _, manualRule := range manualRules {
				if manualRule.Description == key {
					found = true
					teams = manualRule.Teams
				}
			}

			// If rule were not found, return to reprocess the bot comment entirely
			// (maybe bot config was updated since last run?)
			if !found {
				gh.Logger.Debugf("Updated rule not found in config : %s", key)
				return
			}

			// If teams specified in rule, check if actor is a member of one of them
			if len(teams) > 0 {
				if gh.IsUserInTeams(actionCtx.Actor, teams) {
					gh.Logger.Errorf("Checkbox edited by a user not allowed to")
					gh.SetBotComment(previous, int(num))
					os.Exit(1)
				}
			}

			// If box was checked
			reg := regexp.MustCompile(fmt.Sprintf(`(?m:^- \[%s\] %s.*$)`, currentChecks[key][0], key))
			if strings.TrimSpace(currentChecks[key][0]) == "x" {
				replacement := fmt.Sprintf("- [%s] %s (checked by @%s)", currentChecks[key][0], key, actionCtx.Actor)
				edited = reg.ReplaceAllString(current, replacement)
			} else {
				replacement := fmt.Sprintf("- [%s] %s", currentChecks[key][0], key)
				edited = reg.ReplaceAllString(current, replacement)
			}
		}
	}

	// Update comment with username
	if edited != "" {
		gh.SetBotComment(edited, int(num))
		gh.Logger.Debugf("Comment manual checks updated successfully")
	}
}

func updateComment(gh *client.GitHub, pr *github.PullRequest, content CommentContent) {
	// Create bot comment using template file
	const tmplFile = "comment.tmpl"
	tmpl, err := template.New(tmplFile).ParseFiles(tmplFile)
	if err != nil {
		panic(err)
	}

	var commentBytes bytes.Buffer
	if err := tmpl.Execute(&commentBytes, content); err != nil {
		panic(err)
	}

	// Create commit status
	var (
		comment      = gh.SetBotComment(commentBytes.String(), pr.GetNumber())
		context      = "Merge Requirements"
		targetURL    = comment.GetHTMLURL()
		state        = "pending"
		description  = "Some requirements are not satisfied yet. See bot comment."
		allSatisfied = true
	)

	// Check if every requirements are satisfied
	for _, auto := range content.AutoRules {
		if !auto.Satisfied {
			gh.Logger.Infof("AUTO NOT Satisfied", auto.Description)
			allSatisfied = false
			break
		}
	}

	if allSatisfied {
		for _, manual := range content.ManualRules {
			if manual.CheckedBy == "" {
				gh.Logger.Infof("AUTO NOT Satisfied", manual.Description)
				allSatisfied = false
				break
			}
		}
	}

	if allSatisfied {
		state = "success"
		description = "All requirements are satisfied."
	}

	gh.Logger.Infof("STATUS", state, description)

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
		gh.Logger.Errorf("Unable to create status on PR %d : %v", pr.GetNumber(), err)
	}
}
