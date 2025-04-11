package gnominers

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/ownable"
	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"
	"gno.land/r/sys/users"
)

// The user table
type hiddenData struct {
	overallMined uint
	rockMined    uint
	alexandrite  bool
	achievements []iAchievement
}

type UserSession struct {
	hidden    hiddenData
	Money     float64
	GemsCount map[gemType]uint
	Pickaxes  []IPickaxe
	Active    IPickaxe
	Owner     *ownable.Ownable
	DAO       *minerDAO
	Join 	  time.Time
}

func shorten(str string, of int) string {
	return str[:of] + "..." + str[len(str)-of:]
}

func generateCoolName(addr std.Address) string {
	var author string
	user := users.ResolveAddress(addr)
	if user != nil {
		author = "@" + user.Name()
	} else {
		author = shorten(addr.String(), 4)
	}
	return author
}

// Display the user great
func (u *UserSession) GetName() string {
	return generateCoolName(u.Owner.Owner())
}

type UserList []*UserSession

func (a UserList) Len() int           { return len(a) }
func (a UserList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a UserList) Less(i, j int) bool { return true } // idc about less

// The list of users
var userSessions = map[string]*UserSession{}
var userSessionsList = UserList{}

// pub/ The function to log a new user with default values
func Login() {
	caller := std.PreviousRealm().Address()
	infos, ok := userSessions[caller.String()]
	if ok {
		panic("user already registered")
	}
	infos = &UserSession{
		Money:     100,
		GemsCount: map[gemType]uint{},
		Pickaxes:  []IPickaxe{&NormalPickaxe{}},
		Owner:     ownable.NewWithAddress(caller),
		DAO:       nil,
		Join: 	   time.Now(),
	}
	for gem := gemRock; gem < _gemEnd; gem++ {
		infos.GemsCount[gem] = 0
		gem++
	}
	infos.Active = infos.Pickaxes[0]
	userSessions[caller.String()] = infos
	userSessionsList = append(userSessionsList, infos)
	return
}

// sorts
const (
	mostMoney int = iota
	rockMined
	totalMined
	joinedTime
)

// sort names
var sortNames = []string {
	"wealth",
	"rocks mined",
	"total mined",
	"joined time",
}

// functions sort
var sortFunc = []func(x, y *UserSession) bool {
	func(x, y *UserSession) bool {
		return x.Money > y.Money
	},
	func(x, y *UserSession) bool {
		return x.hidden.rockMined > y.hidden.rockMined
	},
	func(x, y *UserSession) bool {
		return x.hidden.overallMined > y.hidden.overallMined
	},
	func(x, y *UserSession) bool {
		return x.Join.Unix() < y.Join.Unix()
	},
}

func sortUsers(slice []*UserSession, less func(i, j *UserSession) bool, rev bool) {
	n := len(slice)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if (!rev && less(slice[j], slice[i])) || (rev && !less(slice[j], slice[i])) {
				slice[i], slice[j] = slice[j], slice[i]
			}
		}
	}
}

// The function to render the user page
func renderUserPage(params map[string]string) string {
	var content string
	lnk := txlink.Call("Login")
	content += "# GnoMiners\n"
	content += "Welcome to gno miners ! Please select an user realm, or [Register](" + lnk + ") if you want to create your own.\n\n"

	// sessions
	content += "<gno-columns>" + br
	content += "## Here are all the sessions:\n"
	if len(userSessions) == 0 {
		content += "Hmmm, looks like there is nobody for now... be the first one to play !\n"
	}
	for k, v := range userSessions {
		content += "- Session of [" + v.GetName() + "](" + gnominersDashboard + "?user=" + k + ")\n"
	}
	content += "\n"
	content += "If you want to play, you can [Register](" + lnk + ")" + br

	// hall of fame
	var sortMethod int
	var err error
	sortStr, ok := params["sort"]
	if !ok {
		sortMethod = mostMoney
	} else {
		sortMethod, err = strconv.Atoi(sortStr)
		if err != nil {
			sortMethod = mostMoney
		}
	}
	_, desc := params["desc"]
	content += "|||" + br
	content += "## Best players\n"
	content += "Sort by: "
	for k, name := range sortNames {
		if k != 0 {
			content += " / "
		}
		baseLink := gnominersPath + "?sort=" + strconv.Itoa(k)
		if k != sortMethod {
			content += md.Link(name, baseLink)
			continue
		}
		if !desc {
			baseLink += "&desc=1"
		}
		content += md.Bold(md.Link(name, baseLink))
		content += " "
		if desc {
			content += "▲"
		} else {
			content += "▼"
		}
	}
	content += br
	cpySessions := make([]*UserSession, len(userSessionsList))
	copy(cpySessions, userSessionsList)
	sortUsers(cpySessions, sortFunc[sortMethod], desc)
	for i := 0; i < len(cpySessions) && i < 5; i++ {
		content += "- " + cpySessions[i].GetName() + "\n"
	}
	content += br
	content += "</gno-columns>" + br
	return content
}
