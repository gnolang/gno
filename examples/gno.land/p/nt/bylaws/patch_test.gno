package bylaws

import (
	"strings"
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

func TestDiffApplyLifecycle(t *testing.T) {
	b := New()

	// Create: diff against an absent document is a create patch.
	create, err := b.Diff("bylaws/quorum.md", "Quorum is half the council.")
	urequire.NoError(t, err)
	uassert.True(t, create.IsCreate(), "expect create patch")
	uassert.False(t, create.IsNoop())
	urequire.NoError(t, b.Apply(create))
	text, _ := b.Get("bylaws/quorum.md")
	uassert.Equal(t, "Quorum is half the council.", text)

	// Amend: a small edit patches to the new text and re-pins the hash.
	amend, err := b.Diff("bylaws/quorum.md", "Quorum is two thirds of the council.")
	urequire.NoError(t, err)
	uassert.False(t, amend.IsCreate())
	uassert.Equal(t, HashText("Quorum is half the council."), amend.Base)
	urequire.NoError(t, b.Apply(amend))
	text, _ = b.Get("bylaws/quorum.md")
	uassert.Equal(t, "Quorum is two thirds of the council.", text)

	// Remove: empty proposed text deletes the document.
	remove, err := b.Diff("bylaws/quorum.md", "")
	urequire.NoError(t, err)
	uassert.True(t, remove.IsRemove(), "expect remove patch")
	urequire.NoError(t, b.Apply(remove))
	uassert.False(t, b.Has("bylaws/quorum.md"))
	uassert.Equal(t, 0, b.Size())
}

func TestApplyStale(t *testing.T) {
	b := New()
	mustCreate(t, b, "a.md", "one")

	// Both patches diff against "one"; the first to apply wins.
	p1, err := b.Diff("a.md", "two")
	urequire.NoError(t, err)
	p2, err := b.Diff("a.md", "three")
	urequire.NoError(t, err)

	urequire.NoError(t, b.Apply(p1))
	uassert.ErrorIs(t, b.Apply(p2), ErrStalePatch)
	text, _ := b.Get("a.md")
	uassert.Equal(t, "two", text) // loser did not clobber

	// A create patch against a now-existing document is stale too.
	create, err := DiffTexts("a.md", "", "fresh", false)
	urequire.NoError(t, err)
	uassert.ErrorIs(t, b.Apply(create), ErrStalePatch)

	// A patch pinning sha256("") is NOT a create: documents are never
	// stored empty, so it can never match anything.
	notCreate := Patch{Path: "b.md", Base: HashText(""), Ops: []Op{{Type: OpInsert, Text: "x"}}}
	uassert.ErrorIs(t, b.Apply(notCreate), ErrStalePatch)
}

func TestApplyRejectsBadScripts(t *testing.T) {
	b := New()
	mustCreate(t, b, "a.md", "abcdef")
	base := b.Hash("a.md")

	cases := []struct {
		name string
		ops  []Op
	}{
		{"keep overruns the base", []Op{{Type: OpKeep, N: 7}}},
		{"delete overruns the base", []Op{{Type: OpDelete, N: 7}}},
		{"script underconsumes the base", []Op{{Type: OpKeep, N: 3}}},
		{"zero keep", []Op{{Type: OpKeep, N: 0}, {Type: OpKeep, N: 6}}},
		{"negative delete", []Op{{Type: OpDelete, N: -1}, {Type: OpKeep, N: 6}}},
		{"empty insert", []Op{{Type: OpKeep, N: 6}, {Type: OpInsert}}},
		{"unknown op", []Op{{Type: OpType('X'), N: 6}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := b.Apply(Patch{Path: "a.md", Base: base, Ops: tc.ops})
			uassert.ErrorIs(t, err, ErrInvalidPatch)
			text, _ := b.Get("a.md")
			uassert.Equal(t, "abcdef", text) // unchanged on error
		})
	}

	// Path and size guards.
	uassert.ErrorIs(t, b.Apply(Patch{Path: "bad path"}), ErrInvalidPath)
	grow := Patch{Path: "a.md", Base: base, Ops: []Op{
		{Type: OpDelete, N: 6},
		{Type: OpInsert, Text: strings.Repeat("x", MaxDocLen+1)},
	}}
	uassert.ErrorIs(t, b.Apply(grow), ErrDocTooLarge)

	_, err := DiffTexts("a.md", "", strings.Repeat("x", MaxDocLen+1), false)
	uassert.ErrorIs(t, err, ErrDocTooLarge)
}

func TestApplyUnicode(t *testing.T) {
	b := New()
	mustCreate(t, b, "a.md", "héllo wörld ✓")

	// Ops count runes, not bytes: a diff over multi-byte text replays
	// exactly.
	p, err := b.Diff("a.md", "héllo brave wörld ✓✓")
	urequire.NoError(t, err)
	urequire.NoError(t, b.Apply(p))
	text, _ := b.Get("a.md")
	uassert.Equal(t, "héllo brave wörld ✓✓", text)
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	b := New()
	mustCreate(t, b, "bylaws/quorum.md", "Quorum is half the council.\nVotes are public.")

	patches := []Patch{}
	p1, err := b.Diff("bylaws/quorum.md", "Quorum is two thirds.\nVotes are public.")
	urequire.NoError(t, err)
	patches = append(patches, p1)
	p2, err := DiffTexts("new/doc.md", "", "Created; with punct: [x] (y) \"z\"\nand a second line.", false)
	urequire.NoError(t, err)
	patches = append(patches, p2)
	p3, err := b.Diff("bylaws/quorum.md", "") // remove
	urequire.NoError(t, err)
	patches = append(patches, p3)

	for _, p := range patches {
		enc := p.Encode()
		dec, err := DecodePatch(enc)
		urequire.NoError(t, err)
		uassert.Equal(t, p.Path, dec.Path)
		uassert.Equal(t, p.Base, dec.Base)
		urequire.Equal(t, len(p.Ops), len(dec.Ops))
		for i := range p.Ops {
			uassert.True(t, p.Ops[i].Type == dec.Ops[i].Type, "op type")
			uassert.Equal(t, p.Ops[i].N, dec.Ops[i].N)
			uassert.Equal(t, p.Ops[i].Text, dec.Ops[i].Text)
		}
		// Deterministic: encoding is stable.
		uassert.Equal(t, enc, dec.Encode())
	}

	// The decoded patch applies like the original.
	dec, err := DecodePatch(p1.Encode())
	urequire.NoError(t, err)
	urequire.NoError(t, b.Apply(dec))
	text, _ := b.Get("bylaws/quorum.md")
	uassert.Equal(t, "Quorum is two thirds.\nVotes are public.", text)
}

func TestDecodePatchRejectsMalformed(t *testing.T) {
	valid, err := DiffTexts("a.md", "", "hello", false)
	urequire.NoError(t, err)
	enc := valid.Encode()
	uassert.Equal(t, "v0:a.md::I5:hello;", enc) // pin the wire format

	bad := []string{
		"",
		"v1:a.md::I5:hello;",           // unknown version
		"v0:bad path::I5:hello;",       // invalid path
		"v0:a.md:zz:I5:hello;",         // malformed base
		"v0:a.md::X5:hello;",           // unknown op
		"v0:a.md::I5:hell",             // truncated literal
		"v0:a.md::I5:helloX",           // missing op terminator
		"v0:a.md::K0;",                 // zero count
		"v0:a.md::K-1;",                // negative count
		"v0:a.md::Kx;",                 // non-numeric count
		"v0:a.md::K05;",                // non-canonical count (leading zero)
		"v0:a.md::K+5;",                // non-canonical count (sign)
		"v0:a.md::I+5:hello;",          // non-canonical insert length
		"v0:a.md::K1",                  // missing terminator
		"v0:a.md::I0:;",                // empty insert
		"v0:a.md",                      // missing base separator
		"v0:a.md::K99999999999999999;", // over MaxDocLen
		"v0:a.md::I1:\xc3;",            // insert is not valid UTF-8
	}
	for _, s := range bad {
		_, err := DecodePatch(s)
		uassert.Error(t, err, "expect decode error: "+s)
	}
}

func TestDecodePatchMaxOps(t *testing.T) {
	// Exactly MaxOps ops decode; one more is rejected. (Shape-only: the
	// script need not fit any document to decode.)
	payload := "v0:a.md::" + strings.Repeat("K1;", MaxOps)
	_, err := DecodePatch(payload)
	uassert.NoError(t, err)

	_, err = DecodePatch(payload + "K1;")
	uassert.ErrorIs(t, err, ErrInvalidPatch)
}

func TestEncodeDecodeMultibyteInsert(t *testing.T) {
	// The insert length prefix is a BYTE length: pin the exact wire form
	// for a multi-byte literal so a rune-length regression cannot pass.
	p, err := DiffTexts("a.md", "", "héllo ✓ wörld", false)
	urequire.NoError(t, err)
	enc := p.Encode()
	uassert.Equal(t, "v0:a.md::I17:héllo ✓ wörld;", enc)

	dec, err := DecodePatch(enc)
	urequire.NoError(t, err)
	b := New()
	urequire.NoError(t, b.Apply(dec))
	text, _ := b.Get("a.md")
	uassert.Equal(t, "héllo ✓ wörld", text)
}

func TestDiffTextsRejectsInvalidUTF8(t *testing.T) {
	_, err := DiffTexts("a.md", "", "bad \xc3 byte", false)
	uassert.ErrorIs(t, err, ErrInvalidText)

	_, err = DiffTexts("a.md", "bad \xc3 base", "fine", true)
	uassert.ErrorIs(t, err, ErrInvalidText)
}

func TestFormat(t *testing.T) {
	b := New()
	mustCreate(t, b, "a.md", "keep DELETE keep")

	p, err := b.Diff("a.md", "keep INSERT keep")
	urequire.NoError(t, err)
	out, err := p.Format("keep DELETE keep")
	urequire.NoError(t, err)
	uassert.True(t, strings.Contains(out, "unchanged"), "expect kept marker")

	// The formatted change shows what leaves and what enters.
	uassert.True(t, strings.Contains(out, "-"), "expect deletion marker")
	uassert.True(t, strings.Contains(out, "+"), "expect insertion marker")

	// Format fails like replay on a script that does not fit the base by
	// rune count. (Fit is positional — content verification is the base
	// hash's job, checked by Apply, not Format.)
	_, err = p.Format("short")
	uassert.ErrorIs(t, err, ErrInvalidPatch)
}

func TestDiffPrefixSuffixTrim(t *testing.T) {
	b := New()
	mustCreate(t, b, "a.md", "The quick brown fox jumps over the lazy dog.")

	// A small middle edit yields Keep(prefix), the middle change, and
	// Keep(suffix): the payload carries only the changed region.
	p, err := b.Diff("a.md", "The quick red fox jumps over the lazy dog.")
	urequire.NoError(t, err)
	urequire.True(t, len(p.Ops) >= 3, "expect trimmed script")
	uassert.True(t, p.Ops[0].Type == OpKeep, "expect leading keep")
	uassert.True(t, p.Ops[len(p.Ops)-1].Type == OpKeep, "expect trailing keep")
	urequire.NoError(t, b.Apply(p))
	text, _ := b.Get("a.md")
	uassert.Equal(t, "The quick red fox jumps over the lazy dog.", text)
}

func TestDiffLargeRewriteFallsBackToReplace(t *testing.T) {
	// Two texts with a shared prefix/suffix but a differing middle beyond
	// the Myers budget: Diff must fall back to a whole-middle replace
	// (bounded ops, no quadratic diff) that still replays exactly.
	base := "HEAD\n" + strings.Repeat("ab", 600) + "\nTAIL"
	proposed := "HEAD\n" + strings.Repeat("cd", 600) + "\nTAIL"

	p, err := DiffTexts("big.md", base, proposed, true)
	urequire.NoError(t, err)
	urequire.True(t, len(p.Ops) <= 4, "expect fallback replace, got ops count")

	b := New()
	create, err := b.Diff("big.md", base)
	urequire.NoError(t, err)
	urequire.NoError(t, b.Apply(create))
	urequire.NoError(t, b.Apply(p))
	text, _ := b.Get("big.md")
	uassert.Equal(t, proposed, text)
}

func TestFormatMarkerSpoof(t *testing.T) {
	// A multi-line insert literal gets every line marker-prefixed, so an
	// embedded "- fake" line cannot masquerade as a deletion marker.
	p, err := DiffTexts("a.md", "", "x\n- fake deletion", false)
	urequire.NoError(t, err)
	out, err := p.Format("")
	urequire.NoError(t, err)
	uassert.True(t, strings.Contains(out, "+ x\n+ - fake deletion\n"), "expect per-line insert markers")
	uassert.False(t, strings.Contains(out, "\n- fake deletion"), "expect no bare deletion-marker line")
}

func TestNoopPatch(t *testing.T) {
	b := New()
	mustCreate(t, b, "a.md", "same")

	p, err := b.Diff("a.md", "same")
	urequire.NoError(t, err)
	uassert.True(t, p.IsNoop(), "expect noop patch")
	urequire.NoError(t, b.Apply(p)) // harmless
	text, _ := b.Get("a.md")
	uassert.Equal(t, "same", text)

	// Creating an empty document is a noop create: nothing is stored.
	pc, err := DiffTexts("empty.md", "", "", false)
	urequire.NoError(t, err)
	uassert.True(t, pc.IsNoop(), "expect noop create")
	urequire.NoError(t, b.Apply(pc))
	uassert.False(t, b.Has("empty.md"))
}
