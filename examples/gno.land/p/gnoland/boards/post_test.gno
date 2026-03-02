package boards_test

import (
	"strings"
	"testing"

	"gno.land/p/nt/urequire/v0"

	"gno.land/p/gnoland/boards"
)

func TestPostSummary(t *testing.T) {
	post := &boards.Post{ID: 1, Body: strings.Repeat("X", 900)}
	summary := post.Summary()
	urequire.True(t, strings.HasSuffix(summary, "..."), "expect dotted suffix")
	urequire.True(t, len(summary) == 80, "expect summary length to match")
}

func TestIsThread(t *testing.T) {
	post := &boards.Post{ID: 1, ThreadID: 1} // IDs match
	urequire.True(t, boards.IsThread(post), "expect post to be a thread")
	urequire.False(t, boards.IsThread(nil), "expect nil not to be a thread")

	post = &boards.Post{ID: 2, ThreadID: 1} // IDs doesn't match
	urequire.False(t, boards.IsThread(post), "expect post not to be a thread")
}

func TestIsRepost(t *testing.T) {
	post := &boards.Post{ID: 1, OriginalBoardID: 1} // Original board ID available
	urequire.True(t, boards.IsRepost(post), "expect post to be a repost")
	urequire.False(t, boards.IsRepost(nil), "expect nil not to be a repost")

	post = &boards.Post{ID: 1} // Original board ID not available
	urequire.False(t, boards.IsRepost(post), "expect post not to be a repost")
}

func TestSummaryOf(t *testing.T) {
	summary := boards.SummaryOf(strings.Repeat("X", 90), 80)
	urequire.True(t, strings.HasSuffix(summary, "..."), "expect dotted suffix")
	urequire.True(t, len(summary) == 80, "expect summary length to match")

	summary = boards.SummaryOf(strings.Repeat(" ", 90), 80)
	urequire.Empty(t, summary, "expect summary to be empty")
}
