package gnominers

import (
	"strconv"

	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/txlink"
	"gno.land/p/pol/dao"
	"gno.land/p/pol/nfmt"
)

const br = "\n\n"

// pub/ Mine some gems with your active pickaxe
func Mine() {
	infos := assertCallerExists()
	infos.Owner.AssertCallerIsOwner()
	for i := 0; i < int(infos.Active.HowMuchMining()); i++ {
		gem, _ := randomGem()
		infos.GemsCount[gem] += 1
		infos.hidden.overallMined += 1
		switch gem {
		case gemRock:
			infos.hidden.rockMined += 1
			achievements["1000rocks"].Check(infos)
		case gemAlexandrite:
			infos.hidden.alexandrite = true
			achievements["alexandrite"].Check(infos)
		}
	}
	achievements["a gem"].Check(infos)
	achievements["diamonds"].Check(infos)
	achievements["42k gems"].Check(infos)
}

// pub/ The function to equip a pickaxe
func Equip(index string) {
	infos := assertCallerExists()
	i, err := strconv.Atoi(index)
	if err != nil {
		return
	}
	infos.Active = infos.Pickaxes[i]
}

// pub/ Open a new DAO
func OpenDAO() {
	infos := assertCallerExists()
	infos.Owner.AssertCallerIsOwner()
	infos.DAO = newMinerDAO()
	infos.DAO.AddMember(dao.Member(infos.Owner.Owner()))
	achievements["dao"].Check(infos)
}

// Render the Dashboard
func renderDashboard(params map[string]string) string {
	// Assertions
	user, ok := params["user"]
	if !ok {
		return "user is not specified. [Go back to home](" + gnominersPath + ") to select a user."
	}
	session, ok := userSessions[user]
	if !ok {
		return "this user is not registered. [Go back to home](" + gnominersPath + ") to register."
	}

	// the content
	content := md.H1("Dashboard of " + session.GetName())

	// actions
	content += "<gno-columns>" + br
	content += md.H3("Actions")
	lnk := txlink.Call("Mine")
	content += md.Link("Mine", lnk) + " - "
	content += md.Link("Shop", gnominersShopPath + "?user=" + user) + " - "
	content += md.Link("Achievements", gnominersAchievements + "?user=" + user) + " - "
	if session.DAO == nil {
		lnk = txlink.Call("OpenDAO")
		content += md.Link("Open a new DAO", lnk)
	} else {
		content += md.Link("Your DAO", gnominersDAO + "?user=" + user)
	}
	content += br
	
	// money money money, must be funny
	content += "|||" + br
	content += md.H3("Money")
	x, err := nfmt.FormatRound(session.Money)
	precise := strconv.FormatFloat(session.Money, 'f', 2, 64)
	if err != nil {
		x = precise
	}
	content += "Currently, you have " + x + " GnoCoins." + br
	if err == nil {
		content += "(precisely " + strconv.FormatFloat(session.Money, 'f', 2, 64) + " GnoCoins)" + br
	}
	content += "</gno-columns>" + br
	
	// your gems
	t := mdtable.Table{ Headers: []string{ "Gem", "Quantity", "Price" } }
	content += "<gno-columns>" + br
	content += md.H3("Your Gems")
	var gem gemType = gemRock
	for gem < _gemEnd {
		count, ok := session.GemsCount[gem]
		if !ok {
			count = 0
		}
		n := float64(count)*float64(gemPriceCaratTable[gem])
		x, err := nfmt.FormatRound(n)
		if err != nil {
			x = strconv.FormatFloat(n, 'f', 2, 64)
		}
		t.Append([]string{ gemNames[gem], strconv.Itoa(int(count)), x + " GC" })
		gem++
	}
	content += t.String()
	content += "|||" + br

	// your inventory
	t = mdtable.Table{ Headers: []string{ "Pickaxe", "Equip" } }
	content += md.H3("Your inventory")
	for index, pick := range session.Pickaxes {
		str := ""
		if pick != session.Active {
			tx := txlink.Call("Equip", "index", strconv.Itoa(index))
			str = md.Link("Equip", tx)
		}
		t.Append([]string{pick.String(), str})
	}
	content += t.String()
	content += "</gno-columns>"
	return content
}
