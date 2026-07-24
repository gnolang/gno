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
	out += "Admin: `" + admin.String() + "`\n\n"
	out += "Message: " + escapeRenderText(message) + "\n"
	if path != "" {
		out += "\nPath: `" + escapeRenderText(path) + "`\n"
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
