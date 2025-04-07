package gnominers

import (
	"strconv"
	"time"

	"gno.land/p/moul/txlink"
)

const shopImage = "https://raw.githubusercontent.com/paulogarithm/gnoplayground/refs/heads/main/images/gnoshop.png"

const birthday = "09/01"
const myAdress = "g1euj8ry6ygx02x48u4505heckkclyps3w87uqtj"

var avaiblePickaxes = map[IPickaxe]float64 {
	&TriPickaxe{}: 110.0,
	&PentaPickaxe{}: 1000.0,
	&DecaPickaxe{}: 10000.42,
	&BirthdayPickaxe{}: 0xbebe_caca, // baby poop in french
}

func totalGemsPrice(session *UserInfo) (total float64) {
	for k, v := range session.GemsCount {
		total += gemPriceCaratTable[k] * float64(v)
	}
	total *= 100
	total = float64(uint(total))
	total /= 100
	return
}

func Sell() {
	session := assertCallerExists()
	session.Owner.AssertCallerIsOwner()
	money := totalGemsPrice(session)
	session.GemsCount = map[gemType]uint{}
	session.Money += money
	achievements["million"].Check(session)
}

func Buy(pickname string) {
	session := assertCallerExists()
	session.Owner.AssertCallerIsOwner()

	// Check for birthday
	birthdayPick := BirthdayPickaxe{}
	today := time.Now().Format("02/01")
	if pickname == birthdayPick.String() && today != birthday {
		panic("its not my birthday yet")
	}

	// Get the pickaxe
	var pick IPickaxe = nil
	price := 0.
	for k, v := range avaiblePickaxes {
		if k.String() == pickname {
			pick = k
			price = v
			break
		}
	}
	if pick == nil {
		panic("invalid pickaxe name " + pickname)
	}

	// Buy if you have the money (unless MY address is trying to buy MY pickaxe)
	if pickname == birthdayPick.String() && session.Adress.String() == myAdress {
		goto addPickaxeAnyway
	}
	if session.Money < price {
		panic("youre too broke to buy this")
	}
	session.Money -= price
	addPickaxeAnyway:
	session.Pickaxes = append(session.Pickaxes, pick.Clone())
	achievements["bday"].Check(session)
}

func renderShop(params map[string]string) string {
	// Assertions
	user, ok := params["user"]
	if !ok {
		return "user is not specified. [Go back to home](" + gnominersPath + ") to select a user."
	}
	session, ok := userSessions[user]
	if !ok {
		return "this user is not registered. [Go back to home](" + gnominersPath + ") to register."
	}

	// Main shop content
	content := "# Shop\n"
	content += "Go back to [DashBoard](" + gnominersDashboard + "?user=" + user + ")."
	content += "![Alt text](" + shopImage + " \"Optional title\")\n\n"

	// Sell
	content += "## Actions\n"
	sellLink := txlink.Call("Sell")
	content += "### [Sell](" + sellLink + ") your rocks"
	content += " (worth " + strconv.FormatFloat(totalGemsPrice(session), 'f', 2, 64) + " GnoCoins).\n\n\n"
	content += "or...\n\n"

	// Buy
	content += "### Buy some pickaxes (you have " + strconv.FormatFloat(session.Money, 'f', 2, 64) + " GnoCoins):\n\n"
	content += "|Pickaxe|Price|Buy|\n|-|-|-|\n"
	birthdayPick := BirthdayPickaxe{}
	today := time.Now().Format("02/01")
	for pick2buy, price := range avaiblePickaxes {
		havePick := false
		for _, userPick := range session.Pickaxes {
			if userPick.String() == pick2buy.String() {
				havePick = true
				break
			}
		}
		if pick2buy.String() == birthdayPick.String() && today != birthday {
			continue
		}
		if !havePick {
			buyLink := txlink.Call("Buy", "pickname", pick2buy.String())
			content += "|" + pick2buy.String()
			content += "|" + strconv.FormatFloat(price, 'f', -1, 64)
			content += "|[Buy](" + buyLink + ")|\n"
		}
	}
	content += "\n"
	return content
}