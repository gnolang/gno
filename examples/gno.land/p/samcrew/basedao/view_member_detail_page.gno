package basedao

import "gno.land/p/nt/ufmt"

func (d *DAOPrivate) MemberDetailPageView(address string) string {
	return d.MemberDetailHeaderView() +
		d.MemberDetailView(address)
}

func (d *DAOPrivate) MemberDetailHeaderView() string {
	name := d.GetProfileString(d.Realm.Address(), "DisplayName", "DAO")
	s := ""
	s += ufmt.Sprintf("# %s - Member Detail\n\n", name)
	s += ufmt.Sprintf("\n--------------------------------\n")
	return s
}

func (d *DAOPrivate) MemberDetailView(addr string) string {
	pkgPath := d.Realm.PkgPath()
	linkPath := getLinkPath(pkgPath)
	roles := d.Members.GetMemberRoles(addr)
	displayName := d.GetProfileString(address(addr), "DisplayName", FALLBACK_DISPLAY_NAME)
	bio := d.GetProfileString(address(addr), "Bio", "No bio")
	pp := d.GetProfileString(address(addr), "Avatar", "")
	s := ""
	s += ufmt.Sprintf("## Profile 👤\n\n")
	s += ufmt.Sprintf("- **Display Name:** %s\n\n", displayName)
	s += ufmt.Sprintf("- **Bio:** %s\n\n", bio)
	if pp != "" {
		s += ufmt.Sprintf("![Avatar](%s)\n\n", pp)
	}
	s += ufmt.Sprintf("## Roles 🎭\n\n")
	for _, role := range roles {
		s += ufmt.Sprintf("- %s\n\n", d.RenderRoleLinkWithChip(role))
	}
	s += ufmt.Sprintf("> Learn more about the roles on the [configuration page ⚙️](%s:%s)\n\n", linkPath, CONFIG_PATH)
	s += ufmt.Sprintf("\n--------------------------------\n")
	return s
}
