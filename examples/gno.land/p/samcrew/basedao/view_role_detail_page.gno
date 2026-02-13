package basedao

import (
	"gno.land/p/demo/svg"
	"gno.land/p/moul/md"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/ufmt"
)

func (d *DAOPrivate) RoleDetailPageView(path string, name string) string {
	return d.RoleDetailHeaderView() +
		d.RoleDetailView(path, name)
}

func (d *DAOPrivate) RoleDetailHeaderView() string {
	name := d.GetProfileString(d.Realm.Address(), "DisplayName", "DAO")
	s := ""
	s += md.H1(ufmt.Sprintf("%s - Role Detail 🎭", name))
	s += ufmt.Sprintf("\n--------------------------------\n")
	return s
}

func (d *DAOPrivate) RoleDetailView(path string, name string) string {
	info := d.Members.RoleInfo(name)
	memberCount := d.Members.CountMembersWithRole(name)
	roleMembers := d.Members.GetMembersWithRole(name)
	var roleMembersTree = avl.NewTree()
	for _, member := range roleMembers {
		roleMembersTree.Set(member, struct{}{})
	}

	s := ""
	s += md.H2(ufmt.Sprintf("Name - %s 📜", info.Name))
	s += md.H2(ufmt.Sprintf("Color - %s %s", info.Color, d.RoleColoredChip(name, &svg.Canvas{
		Width:  24,
		Height: 24,
	})))
	s += md.H2("Description 📝")
	s += md.Paragraph(info.Description)
	s += md.H2(ufmt.Sprintf("Members 👥 (%d)", memberCount))
	s += d.RenderMembersTable(path, roleMembersTree)
	s += ufmt.Sprintf("\n--------------------------------\n")

	return s
}
