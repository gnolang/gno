// Package securitypatterns demonstrates compact realm-authoring security patterns.
package securitypatterns

import "strings"

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
	out += "Message: " + escapeMarkdown(message) + "\n"
	if path != "" {
		out += "\nPath: `" + escapeMarkdown(path) + "`\n"
	}
	return out
}

func assertAdmin(cur realm) bool {
	if !cur.IsCurrent() {
		panic("invalid realm")
	}
	if cur.Previous().Address() != admin {
		panic("admin only")
	}
	return true
}

func escapeMarkdown(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "`", "\\`")
	s = strings.ReplaceAll(s, "[", "\\[")
	s = strings.ReplaceAll(s, "]", "\\]")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
