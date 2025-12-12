package basedao

import (
	"strings"
	"time"

	"gno.land/p/demo/svg"
	"gno.land/p/moul/md"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/avl/pager"
	"gno.land/p/nt/seqid"
	"gno.land/p/nt/ufmt"
	"gno.land/p/samcrew/daokit"
)

// Truncates a string to a maximum length, adding "..." in the middle if it exceeds the limit.
func TruncateMiddle(s string, max ...int) string {
	defMax := 12
	if len(max) > 0 {
		defMax = max[0]
	}
	runes := []rune(s)
	if len(runes) <= defMax {
		return s
	}
	if defMax <= 3 {
		return string(runes[:defMax])
	}
	half := (defMax - 3) / 2
	return ufmt.Sprintf("%s...%s", string(runes[:half]), string(runes[len(runes)-half:]))
}

// Formats an address as a clickable link in Markdown.
func RenderAddressLink(addr string) string {
	return md.Link(TruncateMiddle(addr), ufmt.Sprintf("/u/%s", addr))
}

// Replaces all occurrences of roles in the given string with clickable links in Markdown.
func (d *DAOPrivate) RenderWithRolesLinks(s string) string {
	roles := d.Members.GetRoles()
	for _, role := range roles {
		s = strings.ReplaceAll(s, role, d.RenderRoleLink(role))
	}
	return s
}

// Formats a role as a clickable link in Markdown.
func (d *DAOPrivate) RenderRoleLink(role string) string {
	pkgPath := d.Realm.PkgPath()
	linkPath := getLinkPath(pkgPath)
	roleName := d.Members.RoleInfo(role).Name

	if role == "" {
		return "*No role assigned*"
	}
	return ufmt.Sprintf("[%s](%s:%s/%s)", roleName, linkPath, "role", role)
}

// Formats a role as a clickable link in Markdown, with colored chip at left.
func (d *DAOPrivate) RenderRoleLinkWithChip(role string) string {
	pkgPath := d.Realm.PkgPath()
	linkPath := getLinkPath(pkgPath)

	if role == "" {
		return ufmt.Sprintf("%s %s", d.RoleColoredChip(""), "*No role assigned*")
	}
	return ufmt.Sprintf("[%s %s](%s:%s/%s)", d.RoleColoredChip(role), role, linkPath, "role", role)
}

// Formats a time.Time object as a string with UTC offset.
func FormatDateTimeWithUTCOffset(t time.Time) string {
	_, offsetSeconds := t.Zone()
	hours := offsetSeconds / 3600
	minutes := (offsetSeconds % 3600) / 60
	sign := "+"
	if hours < 0 || minutes < 0 {
		sign = "-"
		hours = -hours
		minutes = -minutes
	}
	return ufmt.Sprintf("%s UTC%s%02d:%02d",
		t.Format("2006-01-02 15:04:05"),
		sign, hours, minutes)
}

// Returns a colored chip svg for the given role. If no role is given, a gray chip is returned.
// It can optionally take a custom svg.Canvas to define the size and style of the chip.
func (d *DAOPrivate) RoleColoredChip(role string, canvas ...*svg.Canvas) string {
	roleName := ""
	chipColor := ""
	canvasWidth := 16
	canvasHeight := 16
	canvasElem := []svg.Elem{}
	canvasStyle := avl.NewTree()

	if canvas != nil && len(canvas) > 0 {
		canvasWidth = canvas[0].Width
		canvasHeight = canvas[0].Height
		canvasElem = canvas[0].Elems
		canvasStyle = canvas[0].Style
	}

	if role == "" {
		chipColor = "#CDCDCD"
	} else {
		info := d.Members.RoleInfo(role)
		chipColor = info.Color
		roleName = info.Name
	}
	chipCanvas := &svg.Canvas{
		Width:  canvasWidth,
		Height: canvasHeight,
		Elems:  canvasElem,
		Style:  canvasStyle,
	}

	chipCanvas.Append(svg.NewRectangle(0, 0, canvasWidth, canvasHeight, chipColor))
	return chipCanvas.Render(roleName + " colored chip")
}

// Renders a table of members with their display name, address, roles, and profile link.
// It supports pagination.
func (d *DAOPrivate) RenderMembersTable(path string, members *avl.Tree) string {
	pkgPath := d.Realm.PkgPath()
	linkPath := getLinkPath(pkgPath)
	const pageSize = 10
	pager := pager.NewPager(members, pageSize, false)
	page := pager.MustGetPageByPath(path)

	s := ""
	s += ufmt.Sprintf("| **Name** | **Address 🔗** | **Roles 🎭** | **Profile** |\n")
	s += ufmt.Sprintf("|----------|----------------|--------------|-------------|\n")

	for _, item := range page.Items {
		displayName := d.GetProfileString(address(item.Key), "DisplayName", FALLBACK_DISPLAY_NAME)
		roles := d.Members.GetMemberRoles(item.Key)
		rolesStr := strings.Join(roles, ", ")

		if len(roles) == 0 {
			rolesStr = d.RenderRoleLinkWithChip("")
		}
		for _, role := range roles {
			rolesStr = strings.Replace(rolesStr, role, d.RenderRoleLinkWithChip(role), -1)
		}

		s += ufmt.Sprintf("| %s | %s | %s | [View](%s:%s/%s) |\n",
			displayName,
			RenderAddressLink(item.Key),
			rolesStr,
			linkPath, "member", item.Key)
	}

	s += "\n" + page.Picker(path) + "\n"
	return s
}

func (d *DAOPrivate) RenderProposalsTable(path string, proposals *daokit.ProposalsStore) string {
	pkgPath := d.Realm.PkgPath()
	linkPath := getLinkPath(pkgPath)
	proposalsCount := proposals.Tree.Size()
	const pageSize = 10
	pager := pager.NewPager(proposals.Tree, pageSize, true)
	page := pager.MustGetPageByPath(path)

	s := ""
	if proposalsCount == 0 {
		s += ufmt.Sprintf("\t⚠️ There are no proposals to show\n\n")
	} else {
		s += ufmt.Sprintf("| **ID** | **Resource** | **Proposer** | **CreatedAt** | **Status** |\n")
		s += ufmt.Sprintf("|--------|--------------|--------------|---------------|------------|\n")
	}

	for _, item := range page.Items {
		proposal := item.Value.(*daokit.Proposal)
		key := item.Key
		id, err := seqid.FromString(key)
		if err != nil {
			panic(err)
		}
		resource := d.Core.Resources.Get(proposal.Action.Type())
		s += ufmt.Sprintf("| [%d](%s:%s/%d) | %s | %s | %s | %s |\n",
			uint64(id), linkPath, "proposal", uint64(id),
			resource.DisplayName,
			RenderAddressLink(proposal.ProposerID),
			FormatDateTimeWithUTCOffset(proposal.CreatedAt),
			proposal.Status.String())
	}

	s += "\n" + page.Picker(path) + "\n"
	return s
}
