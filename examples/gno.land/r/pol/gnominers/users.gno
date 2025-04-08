package gnominers

import (
	"std"

	"gno.land/p/demo/ownable"
	"gno.land/p/moul/txlink"
	"gno.land/r/sys/users"
	"gno.land/r/pol/gnominers/pickaxes"
	"gno.land/p/pol/dao"
)

// The user table
type hiddenData struct {
	overallMined uint
	rockMined    uint
	alexandrite  bool
	achievements []iAchievement
}

type UserInfo struct {
	hidden    hiddenData
	Money     float64
	GemsCount map[GemType]uint
	Pickaxes  []pickaxes.IPickaxe
	Active    pickaxes.IPickaxe
	Adress    std.Address
	Owner     *ownable.Ownable
	DAO 	  *dao.DAO
}

func shorten(str string, of int) string {
	return str[:of] + "..." + str[len(str)-of:]
}

// Display the user great
func (u *UserInfo) GetName() string {
	var author string
	user := users.ResolveAddress(u.Adress)
	if user != nil {
		author = "@" + user.Name()
	} else {
		author = shorten(u.Adress.String(), 4)
	}
	return author
}

// The list of users
var userSessions = map[string]*UserInfo{}

// pub/ The function to log a new user with default values
func Login() {
	caller := std.PreviousRealm().Address()
	infos, ok := userSessions[caller.String()]
	if ok {
		panic("user already registered")
	}
	infos = &UserInfo{
		Money:     100,
		GemsCount: map[GemType]uint{},
		Adress:    caller,
		Pickaxes:  []pickaxes.IPickaxe{&pickaxes.NormalPickaxe{}},
		Owner:     ownable.NewWithAddress(caller),
		DAO: 	   nil,
	}
	for gem := GemRock; gem < _GemEnd; gem++ {
		infos.GemsCount[gem] = 0
		gem++
	}
	infos.Active = infos.Pickaxes[0]
	userSessions[caller.String()] = infos
	return
}

// The function to render the user page
func renderUserPage(map[string]string) string {
	var content string
	lnk := txlink.Call("Login")
	content += "# GnoMiners\n"
	content += "Welcome to gno miners ! Please select an user realm, or [Register](" + lnk + ") if you want to create your own.\n\n"
	content += "## Here are all the sessions:\n"
	if len(userSessions) == 0 {
		content += "Hmmm, looks like there is nobody for now... be the first one to play !\n"
	}
	for k, v := range userSessions {
		content += "- Session of [" + v.GetName() + "](" + gnominersDashboard + "?user=" + k + ")\n"
	}
	content += "\n"
	content += "If you want to play, you can [Register](" + lnk + ")\n"
	return content
}
