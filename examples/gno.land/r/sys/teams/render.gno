package teams

import (
	"strings"

	"gno.land/p/nt/ufmt/v0"
)

// Render shows the authority registry. An ACL nobody can read is an ACL nobody
// audits, so this is the first thing a namespace's users will want.
//
//	""      every namespace with a team
//	"<ns>"  one namespace: its members, and whether it has retired the
//	        registered-name path
func Render(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return renderIndex()
	}
	return renderNamespace(path)
}

func renderIndex() string {
	if teams.Size() == 0 {
		return "# Namespace teams\n\nNo namespace has declared a team yet.\n"
	}
	var b strings.Builder
	b.WriteString("# Namespace teams\n\n")
	b.WriteString(ufmt.Sprintf("%d namespace(s).\n\n", teams.Size()))
	b.WriteString("| Namespace | Members | Registered-name path |\n")
	b.WriteString("|---|---:|---|\n")
	teams.Iterate("", "", func(ns string, v any) bool {
		t := v.(*team)
		state := "live"
		if t.retired {
			state = "**retired**"
		}
		b.WriteString(ufmt.Sprintf("| [%s](/r/sys/teams:%s) | %d | %s |\n",
			ns, ns, t.members.Size(), state))
		return false
	})
	return b.String()
}

func renderNamespace(ns string) string {
	t, ok := get(ns)
	if !ok {
		return ufmt.Sprintf("# %s\n\nNo team. Deploy `/r/%s/home` and have it register one.\n", ns, ns)
	}
	var b strings.Builder
	b.WriteString(ufmt.Sprintf("# %s\n\n", ns))
	b.WriteString(ufmt.Sprintf("Declared by `/r/%s/home`.\n\n", ns))
	if t.retired {
		b.WriteString("The registered-name path is **retired**: the address `r/sys/users` resolves for this name can no longer deploy, and only the team below can.\n\n")
	} else {
		b.WriteString("The registered-name path is live: the address `r/sys/users` resolves for this name can still deploy, in addition to the team below.\n\n")
	}
	b.WriteString(ufmt.Sprintf("## Members (%d)\n\n", t.members.Size()))
	if t.members.Size() == 0 {
		b.WriteString("None.\n")
		return b.String()
	}
	t.members.IterateByOffset(0, MaxMembers, func(a address) bool {
		b.WriteString(ufmt.Sprintf("- %s\n", a.String()))
		return false
	})
	return b.String()
}
