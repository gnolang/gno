package bylaws

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"gno.land/p/onbloc/diff"
)

// MaxOps bounds the number of ops a decoded patch may carry. DiffTexts
// output always stays far below it (see maxMyersRunes), so the cap only
// rejects hand-built pathological payloads at the wire boundary.
const MaxOps = 16 * 1024

// maxMyersRunes bounds the combined rune length of the texts handed to
// MyersDiff, whose memory is O((N+M)·D): past the budget, DiffTexts
// falls back to replacing the whole changed region in one delete+insert
// (payload ~ proposed size; replay stays linear). The common prefix and
// suffix are trimmed first, so ordinary human edits — small changes in a
// large document — stay within budget and get a minimal script.
const maxMyersRunes = 1024

// OpType is the kind of a patch operation.
type OpType byte

const (
	OpKeep   OpType = 'K' // keep the next N runes of the base text
	OpDelete OpType = 'D' // delete the next N runes of the base text
	OpInsert OpType = 'I' // insert literal text
)

// Op is one run of a patch's edit script. Keep and Delete address the
// base text positionally (a rune count), so only Insert carries bytes —
// a small edit to a large document stays a small patch.
type Op struct {
	Type OpType
	N    int    // rune count (Keep and Delete only)
	Text string // inserted literal (Insert only)
}

// Patch is a verifiable amendment to one document: the edit script that
// transforms the document's base text into the proposed text, pinned to
// that base by hash. Apply rejects the patch unless the target currently
// hashes to Base, so a patch can never be applied to text it was not
// diffed against.
type Patch struct {
	Path string // target document path
	Base string // hex sha256 of the base text; "" pins "document absent" (create)
	Ops  []Op   // edit script, replayed in order against the base text
}

// IsCreate reports whether the patch creates the document (its base pins
// "document absent").
func (p Patch) IsCreate() bool {
	return p.Base == ""
}

// IsRemove reports whether applying the patch removes the document (the
// script deletes the whole base text and inserts nothing).
func (p Patch) IsRemove() bool {
	if len(p.Ops) == 0 {
		return false
	}
	for _, op := range p.Ops {
		if op.Type != OpDelete {
			return false
		}
	}
	return true
}

// IsNoop reports whether applying the patch leaves the document set
// unchanged (the script only keeps text, or creates an empty document).
// The check is shape-based: a hand-built script that deletes and
// reinserts identical text is not detected (Diff never produces one).
func (p Patch) IsNoop() bool {
	for _, op := range p.Ops {
		if op.Type != OpKeep {
			return false
		}
	}
	return true
}

// Diff builds the patch that changes the document at path to the
// proposed text: it diffs against the document's current text and pins
// its hash. A new path yields a create patch; empty proposed text yields
// a remove patch.
func (b *Bylaws) Diff(path, proposed string) (Patch, error) {
	cur, exists := b.Get(path)
	return DiffTexts(path, cur, proposed, exists)
}

// DiffTexts builds the patch transforming base into proposed for the
// document at path. exists reports whether the document currently exists
// (base must be "" when it does not); a patch built with exists=false
// creates the document. Both texts must be valid UTF-8 — documents are
// plaintext, and the invariant keeps Keep/Delete rune math byte-faithful.
func DiffTexts(path, base, proposed string, exists bool) (Patch, error) {
	if !IsValidPath(path) {
		return Patch{}, ErrInvalidPath
	}
	if len(proposed) > MaxDocLen {
		return Patch{}, ErrDocTooLarge
	}
	if !utf8.ValidString(base) || !utf8.ValidString(proposed) {
		return Patch{}, ErrInvalidText
	}
	baseHash := ""
	if exists {
		baseHash = HashText(base)
	}
	return Patch{
		Path: path,
		Base: baseHash,
		Ops:  diffOps(base, proposed),
	}, nil
}

// diffOps builds the coalesced edit script from base to proposed: trim
// the common prefix and suffix, Myers-diff the differing middles, and
// past the maxMyersRunes budget replace the whole middle instead. The
// trimmed middles differ at both ends (or are empty), so the pieces
// never need merging with the surrounding Keep ops.
func diffOps(base, proposed string) []Op {
	b, p := []rune(base), []rune(proposed)

	pre := 0
	for pre < len(b) && pre < len(p) && b[pre] == p[pre] {
		pre++
	}
	suf := 0
	for suf < len(b)-pre && suf < len(p)-pre && b[len(b)-1-suf] == p[len(p)-1-suf] {
		suf++
	}
	bMid, pMid := b[pre:len(b)-suf], p[pre:len(p)-suf]

	ops := []Op{}
	if pre > 0 {
		ops = append(ops, Op{Type: OpKeep, N: pre})
	}
	if len(bMid)+len(pMid) > maxMyersRunes {
		if len(bMid) > 0 {
			ops = append(ops, Op{Type: OpDelete, N: len(bMid)})
		}
		if len(pMid) > 0 {
			ops = append(ops, Op{Type: OpInsert, Text: string(pMid)})
		}
	} else {
		ops = append(ops, coalesce(diff.MyersDiff(string(bMid), string(pMid)))...)
	}
	if suf > 0 {
		ops = append(ops, Op{Type: OpKeep, N: suf})
	}
	return ops
}

// coalesce collapses a per-rune Myers edit script into run-length ops:
// runs of Keep/Delete become counts, runs of Insert carry their literal.
func coalesce(edits []diff.Edit) []Op {
	ops := []Op{}
	var (
		lit     strings.Builder
		count   int
		curType OpType
		have    bool
	)
	flush := func() {
		if !have {
			return
		}
		if curType == OpInsert {
			ops = append(ops, Op{Type: OpInsert, Text: lit.String()})
		} else {
			ops = append(ops, Op{Type: curType, N: count})
		}
		lit.Reset()
		count = 0
		have = false
	}
	for _, e := range edits {
		var t OpType
		switch e.Type {
		case diff.EditKeep:
			t = OpKeep
		case diff.EditInsert:
			t = OpInsert
		default:
			t = OpDelete
		}
		if !have || t != curType {
			flush()
			curType = t
			have = true
		}
		if t == OpInsert {
			lit.WriteRune(e.Char)
		} else {
			count++
		}
	}
	flush()
	return ops
}

// Apply verifies the patch and amends the document set: the target's
// current text must hash to the patch base ("" base means the document
// must not exist), and the edit script must consume exactly that text.
// An empty result removes the document. Apply is the package's only
// mutation; on any error the set is unchanged.
func (b *Bylaws) Apply(p Patch) error {
	if !IsValidPath(p.Path) {
		return ErrInvalidPath
	}
	cur, exists := b.Get(p.Path)
	curHash := ""
	if exists {
		curHash = HashText(cur)
	}
	if p.Base != curHash {
		return ErrStalePatch
	}
	out, err := replay(cur, p.Ops)
	if err != nil {
		return err
	}
	if len(out) > MaxDocLen {
		return ErrDocTooLarge
	}
	if out == "" {
		b.docs.Remove(p.Path)
		return nil
	}
	b.docs.Set(p.Path, out)
	return nil
}

// replay runs the edit script against the base text: Keep emits base
// runes and advances, Delete advances, Insert emits its literal. The
// script must consume the base exactly, so a script that does not fit
// the text it runs against fails instead of producing garbage.
func replay(base string, ops []Op) (string, error) {
	r := []rune(base)
	var (
		sb strings.Builder
		i  int
	)
	for _, op := range ops {
		switch op.Type {
		case OpKeep:
			if op.N <= 0 || op.N > len(r)-i {
				return "", ErrInvalidPatch
			}
			sb.WriteString(string(r[i : i+op.N]))
			i += op.N
		case OpDelete:
			if op.N <= 0 || op.N > len(r)-i {
				return "", ErrInvalidPatch
			}
			i += op.N
		case OpInsert:
			if op.Text == "" {
				return "", ErrInvalidPatch
			}
			sb.WriteString(op.Text)
		default:
			return "", ErrInvalidPatch
		}
	}
	if i != len(r) {
		return "", ErrInvalidPatch
	}
	return sb.String(), nil
}

// Format renders the patch against its base text as a plain-text change
// summary: kept runs collapse to a marker, deleted and inserted text is
// shown literally with every line marker-prefixed ("- "/"+ "), so a
// multi-line literal cannot masquerade as the summary's own markers (an
// insertion containing "\n- fake" renders as "+ …" and "+ - fake"). It
// fails like replay when the script does not fit the base. The output is
// raw text — callers rendering markdown must escape it (the content is
// document text, and insertions are proposer-controlled).
func (p Patch) Format(base string) (string, error) {
	r := []rune(base)
	var (
		sb strings.Builder
		i  int
	)
	for _, op := range p.Ops {
		switch op.Type {
		case OpKeep:
			if op.N <= 0 || op.N > len(r)-i {
				return "", ErrInvalidPatch
			}
			sb.WriteString("= " + strconv.Itoa(op.N) + " unchanged\n")
			i += op.N
		case OpDelete:
			if op.N <= 0 || op.N > len(r)-i {
				return "", ErrInvalidPatch
			}
			writeMarked(&sb, "- ", string(r[i:i+op.N]))
			i += op.N
		case OpInsert:
			if op.Text == "" {
				return "", ErrInvalidPatch
			}
			writeMarked(&sb, "+ ", op.Text)
		default:
			return "", ErrInvalidPatch
		}
	}
	if i != len(r) {
		return "", ErrInvalidPatch
	}
	return sb.String(), nil
}

// writeMarked writes text with every line prefixed by marker.
func writeMarked(sb *strings.Builder, marker, text string) {
	for _, line := range strings.Split(text, "\n") {
		sb.WriteString(marker)
		sb.WriteString(line)
		sb.WriteByte('\n')
	}
}

// Encode serializes the patch to a compact single-string payload fit for
// a transaction argument: "v0:<path>:<base>:" followed by one segment
// per op — "K<n>;" and "D<n>;" carry rune counts, "I<len>:<bytes>;"
// carries the inserted literal length-prefixed by its byte length (no
// escaping needed). DecodePatch is the exact inverse.
func (p Patch) Encode() string {
	var sb strings.Builder
	sb.WriteString("v0:")
	sb.WriteString(p.Path)
	sb.WriteByte(':')
	sb.WriteString(p.Base)
	sb.WriteByte(':')
	for _, op := range p.Ops {
		switch op.Type {
		case OpKeep, OpDelete:
			sb.WriteByte(byte(op.Type))
			sb.WriteString(strconv.Itoa(op.N))
		case OpInsert:
			sb.WriteByte(byte(OpInsert))
			sb.WriteString(strconv.Itoa(len(op.Text)))
			sb.WriteByte(':')
			sb.WriteString(op.Text)
		}
		sb.WriteByte(';')
	}
	return sb.String()
}

// DecodePatch parses an Encode payload back into a Patch, validating the
// path, the base hash shape, and every op (counts must be canonical, so
// DecodePatch accepts exactly Encode's output shape; insert literals
// must be valid UTF-8, keeping documents plaintext and Keep's rune math
// byte-faithful). Replay validity against the actual document is Apply's
// job; DecodePatch only guarantees the patch is well-formed.
func DecodePatch(s string) (Patch, error) {
	if !strings.HasPrefix(s, "v0:") {
		return Patch{}, ErrInvalidPatch
	}
	rest := s[len("v0:"):]
	i := strings.IndexByte(rest, ':')
	if i < 0 {
		return Patch{}, ErrInvalidPatch
	}
	path := rest[:i]
	rest = rest[i+1:]
	if !IsValidPath(path) {
		return Patch{}, ErrInvalidPath
	}
	j := strings.IndexByte(rest, ':')
	if j < 0 {
		return Patch{}, ErrInvalidPatch
	}
	base := rest[:j]
	rest = rest[j+1:]
	if !isValidBase(base) {
		return Patch{}, ErrInvalidPatch
	}
	ops := []Op{}
	for len(rest) > 0 {
		if len(ops) == MaxOps {
			return Patch{}, ErrInvalidPatch
		}
		t := OpType(rest[0])
		rest = rest[1:]
		switch t {
		case OpKeep, OpDelete:
			k := strings.IndexByte(rest, ';')
			if k < 0 {
				return Patch{}, ErrInvalidPatch
			}
			n, ok := parseCount(rest[:k])
			if !ok {
				return Patch{}, ErrInvalidPatch
			}
			ops = append(ops, Op{Type: t, N: n})
			rest = rest[k+1:]
		case OpInsert:
			k := strings.IndexByte(rest, ':')
			if k < 0 {
				return Patch{}, ErrInvalidPatch
			}
			n, ok := parseCount(rest[:k])
			if !ok {
				return Patch{}, ErrInvalidPatch
			}
			rest = rest[k+1:]
			if len(rest) < n+1 || rest[n] != ';' {
				return Patch{}, ErrInvalidPatch
			}
			text := rest[:n]
			if !utf8.ValidString(text) {
				return Patch{}, ErrInvalidText
			}
			ops = append(ops, Op{Type: OpInsert, Text: text})
			rest = rest[n+1:]
		default:
			return Patch{}, ErrInvalidPatch
		}
	}
	return Patch{Path: path, Base: base, Ops: ops}, nil
}

// parseCount parses a strictly canonical op count: decimal digits with
// no sign and no leading zero (so Encode∘DecodePatch is the identity on
// accepted payloads), in (0, MaxDocLen].
func parseCount(s string) (int, bool) {
	if s == "" || s[0] == '0' || len(s) > 6 {
		return 0, false
	}
	n := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}
	if n > MaxDocLen {
		return 0, false
	}
	return n, true
}

// isValidBase reports whether a base is the empty sentinel (create) or a
// 64-char lowercase hex sha256.
func isValidBase(base string) bool {
	if base == "" {
		return true
	}
	if len(base) != 64 {
		return false
	}
	for i := 0; i < len(base); i++ {
		c := base[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
