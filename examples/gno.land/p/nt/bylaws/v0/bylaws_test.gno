package bylaws

import (
	"strings"
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

func TestIsValidPath(t *testing.T) {
	valid := []string{
		"bylaws.md",
		"bylaws/quorum.md",
		"mandates/treasury/spend-policy.txt",
		"a",
		"A-b_c.d/e",
	}
	for _, p := range valid {
		uassert.True(t, IsValidPath(p), "expect valid: "+p)
	}

	invalid := []string{
		"",
		"/bylaws.md",
		"bylaws.md/",
		"bylaws//quorum.md",
		"bylaws/./quorum.md",
		"bylaws/../quorum.md",
		".",
		"..",
		"bylaws quorum.md", // space
		"bylaws:quorum",    // encoding delimiter
		"bylaws;quorum",
		"byläws.md", // non-ASCII
		"a\nb",
		strings.Repeat("a", MaxPathLen+1),
	}
	for _, p := range invalid {
		uassert.False(t, IsValidPath(p), "expect invalid: "+p)
	}
}

func TestStoreReads(t *testing.T) {
	b := New()

	// Empty set.
	uassert.Equal(t, 0, b.Size())
	uassert.False(t, b.Has("bylaws/a.md"))
	_, ok := b.Get("bylaws/a.md")
	uassert.False(t, ok)
	uassert.Equal(t, "", b.Hash("bylaws/a.md"))
	uassert.Equal(t, 0, len(b.List("")))

	// Documents enter only through Apply.
	mustCreate(t, b, "bylaws/quorum.md", "Quorum is half.")
	mustCreate(t, b, "bylaws/voting.md", "One member one vote.")
	mustCreate(t, b, "mandates/treasury.md", "Spend only by vote.")

	uassert.Equal(t, 3, b.Size())
	text, ok := b.Get("bylaws/quorum.md")
	urequire.True(t, ok, "expect document")
	uassert.Equal(t, "Quorum is half.", text)
	uassert.Equal(t, HashText("Quorum is half."), b.Hash("bylaws/quorum.md"))

	// List is sorted; a prefix scopes to a folder.
	all := b.List("")
	urequire.Equal(t, 3, len(all))
	uassert.Equal(t, "bylaws/quorum.md", all[0])
	uassert.Equal(t, "bylaws/voting.md", all[1])
	uassert.Equal(t, "mandates/treasury.md", all[2])

	folder := b.List("bylaws/")
	urequire.Equal(t, 2, len(folder))
	uassert.Equal(t, "bylaws/quorum.md", folder[0])
	uassert.Equal(t, "bylaws/voting.md", folder[1])

	// The prefix is raw: without the trailing slash it also matches a
	// sibling file sharing the prefix (documented List behavior).
	mustCreate(t, b, "mandates-old.md", "Superseded.")
	raw := b.List("mandates")
	urequire.Equal(t, 2, len(raw))
	uassert.Equal(t, "mandates-old.md", raw[0])
	uassert.Equal(t, "mandates/treasury.md", raw[1])
	uassert.Equal(t, 1, len(b.List("mandates/")))

	// Iterate stops when fn returns true.
	count := 0
	stopped := b.Iterate("", func(path, text string) bool {
		count++
		return true
	})
	uassert.True(t, stopped, "expect stopped iteration")
	uassert.Equal(t, 1, count)
}

func TestHashText(t *testing.T) {
	// sha256("") — the fixed empty-text digest; documents are never stored
	// empty, so the "" base sentinel (absent) can never collide with it.
	uassert.Equal(t,
		"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		HashText(""))
	uassert.NotEqual(t, HashText("a"), HashText("b"))
}

// mustCreate applies a create patch for a new document.
func mustCreate(t *testing.T, b *Bylaws, path, text string) {
	t.Helper()

	p, err := b.Diff(path, text)
	urequire.NoError(t, err)
	urequire.NoError(t, b.Apply(p))
}
