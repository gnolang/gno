package gnominers

import (
	"strconv"

	"gno.land/p/moul/txlink"
)

// pub/ Mine some gems with your active pickaxe
func Mine() {
	infos := assertCallerExists()
	infos.Owner.AssertCallerIsOwner()
	for i := 0; i < int(infos.Active.HowMuchMining()); i++ {
		gem, _ := randomGem()
		infos.GemsCount[gem] += 1
		infos.hidden.overallMined += 1
		switch gem {
		case GemRock:
			infos.hidden.rockMined += 1
			achievements["1000rocks"].Check(infos)
		case GemAlexandrite:
			infos.hidden.alexandrite = true
			achievements["alexandrite"].Check(infos)
		}
	}
	achievements["a gem"].Check(infos)
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
	content := "# Dashboard of " + session.GetName() + "\n"

	// actions
	content += "### Actions\n"
	lnk := txlink.Call("Mine")
	content += "[Mine](" + lnk + ") - "
	content += "[Shop](" + gnominersShopPath + "?user=" + user + ") - "
	content += "[Achievements](" + gnominersAchievements + "?user=" + user + ") - "
	// if session.DAO != nil {
	// 	lnk = txlink.Call("OpenDAO")
	// 	content += "[Open DAO](" + lnk + ")"
	// } else {
	// 	content += "[Your DAO](" + gnominersAchievements
	// }
	content += "\n\n"

	// money money money, must be funny
	content += "### Money\n"
	content += "Currently, you have " + strconv.FormatFloat(session.Money, 'f', 2, 64) + " GnoCoins\n"

	// your gems
	content += "### Your Gems\n"
	content += "|Gem|Quantity|Price|\n"
	content += "|-|-|-|\n"
	var gem GemType = GemRock
	for gem < _GemEnd {
		count, ok := session.GemsCount[gem]
		if !ok {
			count = 0
		}
		content += "|" + GemNames[gem]
		content += "|" + strconv.Itoa(int(count))
		content += "|" + strconv.FormatFloat(float64(count)*float64(gemPriceCaratTable[gem]), 'f', 2, 64)
		content += "|\n"
		gem++
	}
	content += "\n"

	// your inventory
	content += "### Your inventory\n"
	content += "|Pickaxe|Equip|\n"
	content += "|-|-|\n"
	for index, pick := range session.Pickaxes {
		content += "|" + pick.String() + "|"
		if pick != session.Active {
			tx := txlink.Call("Equip", "index", strconv.Itoa(index))
			content += "[equip](" + tx + ")"
		}
		content += "|\n"
	}
	content += "\n"
	return content
}


// DAO RENDERING HERE

// Render the dao !
func renderDao(map[string]string) string {
	return "dao here"
}