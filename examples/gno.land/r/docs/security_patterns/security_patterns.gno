package security_patterns

import "gno.land/p/nt/markdown/sanitize/v0"

var (
	admin   = address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")
	message = "Only the admin can edit this message."
)

func SetMessage(cur realm, next string) {
	if assertAdmin(cur) {
		message = next
	}
}

func TransferAdmin(cur realm, next address) {
	if !next.IsValid() {
		panic("invalid admin")
	}
	if assertAdmin(cur) {
		admin = next
	}
}

func Admin() address {
	return admin
}

func Message() string {
	return message
}

func Render(path string) string {
	out := "# Security Patterns\n\n"
	out += "This realm demonstrates three defensive patterns; read the source " +
		"alongside this page:\n\n"
	out += "1. **Live-realm guard** — `assertAdmin` panics unless " +
		"`cur.IsCurrent()` holds, so a stale realm token captured elsewhere " +
		"cannot mutate state.\n"
	out += "2. **Caller identity via `cur.Previous().Address()`** — the admin " +
		"check reads the immediate caller, not `OriginCaller()`, so an " +
		"intermediary realm cannot pass itself off as the user.\n"
	out += "3. **Sanitized render output** — every value echoed below is run " +
		"through `p/nt/markdown/sanitize` first, so caller-controlled text " +
		"cannot inject markdown or break out of a code span.\n\n"
	// InlineCode wraps in a backtick run wide enough to outscan any backticks
	// in the content, so a backtick in path cannot close the span early — a
	// naive "`" + path + "`" would.
	out += "Admin: " + sanitize.InlineCode(admin.String()) + "\n\n"
	out += "Message: " + escapeRenderText(message) + "\n"
	if path != "" {
		out += "\nPath: " + sanitize.InlineCode(path) + "\n"
	}
	return out
}

// assertAdmin guards a state-mutating call. The caller must pass its own cur
// realm down; assertAdmin is a non-crossing helper, so it does not shift
// PreviousRealm — the crossing happened in the outer function.
func assertAdmin(cur realm) bool {
	if !cur.IsCurrent() {
		panic("invalid realm")
	}
	if cur.Previous().Address() != admin {
		panic("admin only")
	}
	return true
}

func escapeRenderText(s string) string {
	return sanitize.InlineText(s)
}
