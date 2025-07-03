package check

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/internal/client"
	"github.com/gnolang/gno/contribs/github-bot/internal/logger"
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"
	"github.com/google/go-github/v64/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/sethvargo/go-githubactions"
	"github.com/stretchr/testify/assert"
)

func TestGeneratedComment(t *testing.T) {
	t.Parallel()

	autoCheckSuccessLine := regexp.MustCompile(fmt.Sprintf(`(?m:^ %s .+$)`, utils.Success))
	autoCheckFailLine := regexp.MustCompile(fmt.Sprintf(`(?m:^ %s .+$)`, utils.Fail))

	content := CommentContent{}
	autoRules := []AutoContent{
		{Description: "Test automatic 1", Satisfied: false},
		{Description: "Test automatic 2", Satisfied: false},
		{Description: "Test automatic 3", Satisfied: true},
		{Description: "Test automatic 4", Satisfied: true},
		{Description: "Test automatic 5", Satisfied: false},
	}
	manualRules := []ManualContent{
		{Description: "Test manual 1", CheckedBy: "user-1"},
		{Description: "Test manual 2", CheckedBy: ""},
		{Description: "Test manual 3", CheckedBy: ""},
		{Description: "Test manual 4", CheckedBy: "user-4"},
		{Description: "Test manual 5", CheckedBy: "user-5"},
	}

	commentText, err := generateComment(content)
	assert.Nil(t, err, fmt.Sprintf("error is not nil: %v", err))
	assert.True(t, strings.Contains(commentText, "*No automated checks match this pull request.*"), "should contains automated check placeholder")
	assert.True(t, strings.Contains(commentText, "*No manual checks match this pull request.*"), "should contains manual check placeholder")
	assert.True(t, strings.Contains(commentText, "All **Automated Checks** passed. ✅"), "should contains automated checks passed placeholder")

	content.AutoRules = autoRules
	content.AutoAllSatisfied = true
	commentText, err = generateComment(content)
	assert.Nil(t, err, fmt.Sprintf("error is not nil: %v", err))
	assert.False(t, strings.Contains(commentText, "*No automated checks match this pull request.*"), "should not contains automated check placeholder")
	assert.True(t, strings.Contains(commentText, "*No manual checks match this pull request.*"), "should contains manual check placeholder")
	assert.True(t, strings.Contains(commentText, "All **Automated Checks** passed. ✅"), "should contains automated checks passed placeholder")
	assert.Equal(t, 2, len(autoCheckSuccessLine.FindAllStringSubmatch(commentText, -1)), "wrong number of succeeded automatic check")
	assert.Equal(t, 3, len(autoCheckFailLine.FindAllStringSubmatch(commentText, -1)), "wrong number of failed automatic check")

	content.AutoAllSatisfied = false
	commentText, err = generateComment(content)
	assert.Nil(t, err, fmt.Sprintf("error is not nil: %v", err))
	assert.False(t, strings.Contains(commentText, "*No automated checks match this pull request.*"), "should not contains automated check placeholder")
	assert.True(t, strings.Contains(commentText, "*No manual checks match this pull request.*"), "should contains manual check placeholder")
	assert.False(t, strings.Contains(commentText, "All **Automated Checks** passed. ✅"), "should contains automated checks passed placeholder")
	assert.Equal(t, 2, len(autoCheckSuccessLine.FindAllStringSubmatch(commentText, -1)), "wrong number of succeeded automatic check")
	assert.Equal(t, 3+3, len(autoCheckFailLine.FindAllStringSubmatch(commentText, -1)), "wrong number of failed automatic check")

	content.ManualRules = manualRules
	commentText, err = generateComment(content)
	assert.Nil(t, err, fmt.Sprintf("error is not nil: %v", err))
	assert.False(t, strings.Contains(commentText, "*No automated checks match this pull request.*"), "should not contains automated check placeholder")
	assert.False(t, strings.Contains(commentText, "*No manual checks match this pull request.*"), "should not contains manual check placeholder")
	assert.False(t, strings.Contains(commentText, "All **Automated Checks** passed. ✅"), "should contains automated checks passed placeholder")

	manualChecks := getCommentManualChecks(commentText)
	assert.Equal(t, len(manualChecks), len(manualRules), "wrong number of manual checks found")
	for _, rule := range manualRules {
		val, ok := manualChecks[rule.Description]
		assert.True(t, ok, "manual check should exist")
		if rule.CheckedBy == "" {
			assert.Equal(t, " ", val.status, "manual rule should not be checked")
		} else {
			assert.Equal(t, "x", val.status, "manual rule should be checked")
		}
		assert.Equal(t, rule.CheckedBy, val.checkedBy, "invalid username found for CheckedBy")
	}
}

func setValue(t *testing.T, m map[string]any, value any, keys ...string) map[string]any {
	t.Helper()

	if len(keys) > 1 {
		currMap, ok := m[keys[0]].(map[string]any)
		if !ok {
			currMap = map[string]any{}
		}
		m[keys[0]] = setValue(t, currMap, value, keys[1:]...)
	} else if len(keys) == 1 {
		m[keys[0]] = value
	}

	return m
}

func TestCommentUpdateHandler(t *testing.T) {
	t.Parallel()

	const (
		user = "user"
		bot  = "bot"
	)
	actionCtx := &githubactions.GitHubContext{
		Event: make(map[string]any),
	}

	mockOptions := []mock.MockBackendOption{}
	newGHClient := func() *client.GitHub {
		return &client.GitHub{
			Client: github.NewClient(mock.NewMockedHTTPClient(mockOptions...)),
			Ctx:    context.Background(),
			Logger: logger.NewNoopLogger(),
		}
	}
	gh := newGHClient()

	// Exit without error because EventName is empty.
	assert.NoError(t, handleCommentUpdate(gh, actionCtx))
	actionCtx.EventName = utils.EventIssueComment

	// Exit with error because Event.action is not set.
	assert.Error(t, handleCommentUpdate(gh, actionCtx))
	actionCtx.Event["action"] = ""

	// Exit without error because Event.action is set but not 'deleted'.
	assert.NoError(t, handleCommentUpdate(gh, actionCtx))
	actionCtx.Event["action"] = "deleted"

	// Exit with error because Event.issue.number is not set.
	assert.Error(t, handleCommentUpdate(gh, actionCtx))
	actionCtx.Event = setValue(t, actionCtx.Event, float64(42), "issue", "number")

	// Exit without error can't get open pull request associated with PR num.
	assert.NoError(t, handleCommentUpdate(gh, actionCtx))
	mockOptions = append(mockOptions, mock.WithRequestMatchPages(
		mock.EndpointPattern{
			Pattern: "/repos/pulls/42",
			Method:  "GET",
		},
		github.PullRequest{Number: github.Int(42), State: github.String(utils.PRStateOpen)},
	))
	gh = newGHClient()

	// Exit with error because mock not setup to return authUser.
	assert.Error(t, handleCommentUpdate(gh, actionCtx))
	mockOptions = append(mockOptions, mock.WithRequestMatchPages(
		mock.EndpointPattern{
			Pattern: "/user",
			Method:  "GET",
		},
		github.User{Login: github.String(bot)},
	))
	gh = newGHClient()
	actionCtx.Actor = bot

	// Exit with error because authUser and action actor is the same user.
	assert.ErrorIs(t, handleCommentUpdate(gh, actionCtx), errTriggeredByBot)
	actionCtx.Actor = user

	// Exit with error because Event.comment.user.login is not set.
	assert.Error(t, handleCommentUpdate(gh, actionCtx))
	actionCtx.Event = setValue(t, actionCtx.Event, user, "comment", "user", "login")

	// Exit without error because comment author is not the bot.
	assert.NoError(t, handleCommentUpdate(gh, actionCtx))
	actionCtx.Event = setValue(t, actionCtx.Event, bot, "comment", "user", "login")

	// Exit with error because Event.comment.body is not set.
	assert.Error(t, handleCommentUpdate(gh, actionCtx))
	actionCtx.Event = setValue(t, actionCtx.Event, "current_body", "comment", "body")

	// Exit with error because Event.changes.body.from is not set.
	assert.Error(t, handleCommentUpdate(gh, actionCtx))
	actionCtx.Event = setValue(t, actionCtx.Event, "updated_body", "changes", "body", "from")

	// Exit with error because checkboxes are differents.
	assert.Error(t, handleCommentUpdate(gh, actionCtx))
	actionCtx.Event = setValue(t, actionCtx.Event, "current_body", "changes", "body", "from")

	assert.Nil(t, handleCommentUpdate(gh, actionCtx))
}
