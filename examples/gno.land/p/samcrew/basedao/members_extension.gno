package basedao

import "gno.land/p/samcrew/daokit"

type MembersViewExtension interface {
	IsMember(memberId string) bool
}

const MembersViewExtensionPath = "gno.land/p/demo/basedao.MembersView"

func MustGetMembersViewExtension(dao daokit.DAO) MembersViewExtension {
	ext := dao.Extension(MembersViewExtensionPath)
	if ext == nil {
		panic("no such extension: " + MembersViewExtensionPath)
	}
	val, ok := ext.(MembersViewExtension)
	if !ok {
		panic("extension does not conform to interface: " + MembersViewExtensionPath)
	}
	return val
}

type membersViewExtension struct {
	getStore  func() *MembersStore
	queryPath string
}

func (m *membersViewExtension) IsMember(memberId string) bool {
	return m.getStore().IsMember(memberId)
}

var _ MembersViewExtension = (*membersViewExtension)(nil)

func (m *membersViewExtension) Info() daokit.ExtensionInfo {
	return daokit.ExtensionInfo{
		Path:      MembersViewExtensionPath,
		Version:   "1",
		QueryPath: m.queryPath,
	}
}
