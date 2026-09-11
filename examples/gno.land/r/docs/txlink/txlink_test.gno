package txlink

import (
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

func TestPostCapturesCaller(cur realm, t *testing.T) {
	aliceAddr := testutils.TestAddress("alice")
	testing.SetRealm(testing.NewUserRealm(aliceAddr))

	text := "hello there"
	Post(cross(cur), text)

	got := Render("")
	uassert.True(t, strings.Contains(got, text), "expected render to contain the message text")
	uassert.True(t, strings.Contains(got, aliceAddr.String()), "expected render to contain the caller address")
}

func TestPostSanitizesText(cur realm, t *testing.T) {
	aliceAddr := testutils.TestAddress("alice")
	testing.SetRealm(testing.NewUserRealm(aliceAddr))

	// Markdown control characters in caller-supplied text must be escaped so
	// they render as literal text rather than active markdown.
	Post(cross(cur), "**not bold** [not a link](/r/evil)")

	got := Render("")
	uassert.False(t, strings.Contains(got, "**not bold**"), "expected emphasis markers to be escaped")
	uassert.False(t, strings.Contains(got, "[not a link](/r/evil)"), "expected link syntax to be escaped")
}
