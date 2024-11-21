package main

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/internal/utils"
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
		{Description: "Test manual 1", CheckedBy: "user_1"},
		{Description: "Test manual 2", CheckedBy: ""},
		{Description: "Test manual 3", CheckedBy: ""},
		{Description: "Test manual 4", CheckedBy: "user_4"},
		{Description: "Test manual 5", CheckedBy: "user_5"},
	}

	commentText, err := generateComment(content)
	assert.Nil(t, err, fmt.Sprintf("error is not nil: %v", err))
	assert.True(t, strings.Contains(commentText, "*No automated checks match this pull request.*"), "should contains automated check placeholder")
	assert.True(t, strings.Contains(commentText, "*No manual checks match this pull request.*"), "should contains manual check placeholder")

	content.AutoRules = autoRules
	commentText, err = generateComment(content)
	fmt.Println(commentText)
	assert.Nil(t, err, fmt.Sprintf("error is not nil: %v", err))
	assert.False(t, strings.Contains(commentText, "*No automated checks match this pull request.*"), "should not contains automated check placeholder")
	assert.True(t, strings.Contains(commentText, "*No manual checks match this pull request.*"), "should contains manual check placeholder")
	assert.Equal(t, 2, len(autoCheckSuccessLine.FindAllStringSubmatch(commentText, -1)), "wrong number of succeeded automatic check")
	assert.Equal(t, 3, len(autoCheckFailLine.FindAllStringSubmatch(commentText, -1)), "wrong number of failed automatic check")

	content.ManualRules = manualRules
	commentText, err = generateComment(content)
	assert.Nil(t, err, fmt.Sprintf("error is not nil: %v", err))
	assert.False(t, strings.Contains(commentText, "*No automated checks match this pull request.*"), "should not contains automated check placeholder")
	assert.False(t, strings.Contains(commentText, "*No manual checks match this pull request.*"), "should not contains manual check placeholder")

	manualChecks := getCommentManualChecks(commentText)
	assert.Equal(t, len(manualChecks), len(manualRules), "wrong number of manual checks found")
	for _, rule := range manualRules {
		val, ok := manualChecks[rule.Description]
		assert.True(t, ok, "manual check should exist")
		if rule.CheckedBy == "" {
			assert.Equal(t, " ", val[0], "manual rule should not be checked")
		} else {
			assert.Equal(t, "x", val[0], "manual rule should be checked")
		}
		assert.Equal(t, rule.CheckedBy, val[1], "invalid username found for CheckedBy")
	}
}
