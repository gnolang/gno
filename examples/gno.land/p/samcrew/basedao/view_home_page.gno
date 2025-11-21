package basedao

import (
	"gno.land/p/moul/md"
	"gno.land/p/nt/ufmt"
)

func (d *DAOPrivate) HomePageView(path string) string {
	pkgPath := d.Realm.PkgPath()
	linkPath := getLinkPath(pkgPath)
	return d.HomeProfileHeaderView(path == HOME_NO_PROFILE_PATH) +
		md.Link("> Go to Members", linkPath+":"+MEMBERS_PATH) + "\n\n" +
		md.Link("> Go to Proposals", linkPath+":"+PROPOSALS_PATH) + "\n"
}

func (d *DAOPrivate) HomeProfileHeaderView(noprofile bool) string {
	s := ""
	name := d.GetProfileString(d.Realm.Address(), "DisplayName", "DAO")
	description := d.GetProfileString(d.Realm.Address(), "Bio", "Created with daokit")
	pkgPath := d.Realm.PkgPath()
	linkPath := getLinkPath(pkgPath)

	if !noprofile {
		s += ufmt.Sprintf("# %s\n\n", name)
		imageURI := d.GetProfileString(d.Realm.Address(), "Avatar", "")
		if imageURI != "" {
			s += ufmt.Sprintf("![Image](%s)\n\n", imageURI)
		}
		s += ufmt.Sprintf("%s\n\n", description)
	}

	s += ufmt.Sprintf("> Realm address: %s\n\n", d.Realm.Address())
	s += ufmt.Sprintf("Discover more about this DAO on the [configuration page ⚙️](%s:%s)\n\n", linkPath, CONFIG_PATH)
	s += ufmt.Sprintf("\n--------------------------------\n")

	return s
}
