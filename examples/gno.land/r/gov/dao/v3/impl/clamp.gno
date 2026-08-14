package impl

// Attacker-controlled strings are clamped before they reach the sanitizer, so
// escaping costs a bounded constant rather than scaling with input.
//
// Sanitizing runs ~11,310 gas/byte (InlineCode) against ~31 for the raw
// concatenation it replaced, and Render is reachable unauthenticated through
// vm/qrender under maxGasQuery = 3_000_000_000. ExecutorCreationRealm is
// dispatched through the public dao.Executor interface, so a hostile executor
// computes it per call while storing almost nothing. Measured on the real
// render path: a 250KB value costs 2,839,117,770 gas unclamped against
// 18,749,984 clamped. 250KB is just under the cap; about 265KB crosses it
// (3,008,818,394), and past that the page cannot be rendered at all by
// anyone. Choosing the larger number costs the attacker nothing, since the
// value is computed per call and almost nothing is stored. Removing that
// amplification is the point of this file.
//
// It does NOT make the page safe, and nothing here should be read as claiming
// so. The executor's method body runs inside the same query and is unbounded:
// an executor that simply burns CPU before returning a short string still
// renders the proposal page permanently un-queryable, for a few hundred bytes
// of on-chain storage. That predates this change — Render has always called
// ExecutorString() and ExecutorCreationRealm() through the public interface —
// and bounding it needs a gas budget around executor dispatch, not a clamp.
const (
	// A realm path. The longest deployable realm path in examples/ is 35 bytes
	// (gno.land/r/gov/dao/v3/treasury/test), so this leaves about seven times
	// the room anything real needs. A hostile executor returns any length it
	// likes, which is the reason to clamp at all. It is headroom, not a
	// guarantee: the package-path grammar puts no ceiling on how many segments
	// a path may have, so a deeply nested realm could still be cut. That costs
	// a truncation marker on the page and nothing else.
	maxRenderedRealm = 256
	// "execution failed: " plus an executor's error message. Also bounds what
	// govdao.gno stores, so the realm never holds a reason it cannot show.
	maxRenderedReason = 1024
	// A proposal title. The longest in examples/ is about 40 bytes, so this is
	// roughly nine times anything real, and long for a heading. Titles are
	// escaped on both the proposal page and the list page, and the list page
	// escapes one per proposal shown, so this is the bound that keeps a single
	// oversized title from pricing the whole list out of the query cap.
	maxRenderedTitle = 512
)

// Always clamp first — before escaping, and before any other pass over the
// value. Trimming used to run before the clamp, which meant it walked every
// byte the executor returned: 250KB of spaces cost 1,368,719,824 gas to render
// nothing, and 560KB cost 3,048,904,520, past the query cap. Clamping first
// bounds that scan and brought the same 560KB down to 16,657,792.
//
// Never clamp after escaping either. The escapers size their wrapper
// from the string they are handed — InlineCode picks a fence long enough to
// outscan the backticks it can see. Cutting a value that has already been
// escaped can slice the closing fence off and leave the span hanging open,
// which is worse than not clamping at all. Both call sites read
// InlineCode(clampField(...)) for that reason, and the enormous-value case in
// filetests/executor_disclosure_filetest.gno fails if the two are swapped.
//
// clampField cuts s to at most max bytes, backing off to a rune boundary so a
// well-formed multi-byte character is not split, and marks the result so a
// reader can tell it was cut. Input that is already invalid UTF-8 can still
// leave a dangling lead byte; the sanitizer tolerates that. The marker avoids markdown punctuation:
// these values are escaped downstream, and parentheses would come back as
// "\(truncated\)".
func clampField(s string, max int) string {
	if len(s) <= max {
		return s
	}

	end := max
	// UTF-8 continuation bytes are 0b10xxxxxx.
	for end > 0 && s[end]&0xC0 == 0x80 {
		end--
	}

	return s[:end] + "… truncated"
}
