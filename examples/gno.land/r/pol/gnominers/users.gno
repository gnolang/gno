package gnominers

import (
	"std"
	"gno.land/r/sys/users"
	"gno.land/p/demo/ownable"
	"gno.land/p/moul/txlink"
)

// The user table
type hiddenData struct {
	overallMined uint
	rockMined uint
	alexandrite bool
	achievements []IAchievement
}

type UserInfo struct {
	hidden hiddenData
	Money float64
	GemsCount map[gemType]uint
	Pickaxes []IPickaxe
	Active IPickaxe
	Adress std.Address
	Owner *ownable.Ownable
}

func shorten(str string, of int) string {
	return str[:of] + "..." + str[len(str) - of:]
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

// The function to log a new user with default values
func Login() {
	caller := std.PreviousRealm().Address()
	infos, ok := userSessions[caller.String()]
	if ok {
		panic("user already registered")
	}
	infos = &UserInfo{
		Money: 100,
		GemsCount: map[gemType]uint{},
		Adress: caller,
		Pickaxes: []IPickaxe{ &NormalPickaxe{} },
		Owner: ownable.NewWithAddress(caller),
	}
	for gem := gemRock; gem < _gemEnd; gem++ {
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