package basedao

import (
	"gno.land/p/nt/ufmt"
	"gno.land/p/samcrew/piechart"
)

func (d *DAOPrivate) MembersPageView(path string) string {
	memberCount := d.Members.MembersCount()
	membersNoRoleCount := d.Members.CountMembersWithoutRole()
	roles := d.Members.GetRoles()

	pieSlices := []piechart.PieSlice{}

	for _, role := range roles {
		count := d.Members.CountMembersWithRole(role)
		if count == 0 {
			continue
		}
		roleInfo := d.Members.RoleInfo(role)
		pieSlices = append(pieSlices, piechart.PieSlice{
			Value: float64(count),
			Color: roleInfo.Color,
			Label: roleInfo.Name,
		})
	}

	if membersNoRoleCount > 0 {
		pieSlices = append(pieSlices, piechart.PieSlice{
			Value: float64(membersNoRoleCount),
			Color: "#CDCDCD",
			Label: "No role assigned",
		})
	}

	s := ""
	s += ufmt.Sprintf("## Members 👥 (%d)\n\n", memberCount)
	s += d.RenderMembersTable(path, d.Members.Members) + "\n"
	if len(pieSlices) > 1 {
		s += piechart.Render(pieSlices, "Roles distribution:")
	}
	s += ufmt.Sprintf("\n--------------------------------\n")
	return s
}
