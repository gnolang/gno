package home

import (
    "std"
    "sort"

    "gno.land/p/demo/ufmt"

    "gno.land/r/stefann/config"
)


type City struct {
	Name string
	URL  string
}

type Sponsor struct {
	Address std.Address
	Amount  std.Coins
}

var (
    pfp              string
    cities           []City
    currentCityIndex int
    aboutMe          [2]string
    jarLink          string
    maxSponsors      int
    sponsors         []Sponsor
	totalDonated     std.Coins
	totalDonations   int
)

func init() {
    pfp = "https://i.ibb.co/Bc5YNCx/DSC-0095a.jpg"
    cities = []City{
        {Name: "Venice", URL: "https://i.ibb.co/1mcZ7b1/venice.jpg"},
        {Name: "Tokyo", URL: "https://i.ibb.co/wNDJv3H/tokyo.jpg"},
        {Name: "São Paulo", URL: "https://i.ibb.co/yWMq2Sn/sao-paulo.jpg"},
        {Name: "Toronto", URL: "https://i.ibb.co/pb95HJB/toronto.jpg"},
        {Name: "Bangkok", URL: "https://i.ibb.co/pQy3w2g/bangkok.jpg"},
        {Name: "New York", URL: "https://i.ibb.co/6JWLm0h/new-york.jpg"},
        {Name: "Paris", URL: "https://i.ibb.co/q9vf6Hs/paris.jpg"},
        {Name: "Kandersteg", URL: "https://i.ibb.co/60DzywD/kandersteg.jpg"},
        {Name: "Rothenburg", URL: "https://i.ibb.co/cr8d2rQ/rothenburg.jpg"},
        {Name: "Capetown", URL: "https://i.ibb.co/bPGn0v3/capetown.jpg"},
        {Name: "Sydney", URL: "https://i.ibb.co/TBNzqfy/sydney.jpg"},
        {Name: "Oeschinen Lake", URL: "https://i.ibb.co/QJQwp2y/oeschinen-lake.jpg"},
        {Name: "Barra Grande", URL: "https://i.ibb.co/z4RXKc1/barra-grande.jpg"},
        {Name: "London", URL: "https://i.ibb.co/CPGtvgr/london.jpg"},
    }
    currentCityIndex = 0
    jarLink = "https://TODO"
    maxSponsors = 5
    aboutMe = [2]string{
        `<h3 style="font-size: 1.4em;">About Me</h3>
        <p style="font-size: 1.1em;">Hey there! I’m Stefan, a student of Computer Science. I’m all about exploring and adventure — whether it’s diving into the latest tech or discovering a new city, I’m always up for the challenge!</p>`,
        `<h3 style="font-size: 1.4em;">Contributions</h3>
        <p style="font-size: 1.1em;">I'm just getting started, but you can follow my journey through Gno.land right here <a href="https://github.com/gnolang/hackerspace/issues/94" target="_blank">🔗</a></p>`,
    }
}

func UpdateMaxSponsors(newMax int) {
    config.AssertAuthorized()
    maxSponsors = newMax
}

func UpdateCities(newCities []City) {
	config.AssertAuthorized()
	cities = newCities
}

func UpdateJarLink(newLink string) {
	config.AssertAuthorized()
	jarLink = newLink
}

func UpdatePFP(url, caption string) {
	config.AssertAuthorized()
	pfp = url
}

func UpdateAboutMe(col1, col2 string) {
	config.AssertAuthorized()
	aboutMe[0] = col1
	aboutMe[1] = col2
}

func Donate() {
    address := std.GetOrigCaller()
    amount := std.GetOrigSend()
    
    if amount.AmountOf("ugnot") == 0 {
        panic("Donation must include GNOT")
    }

    found := false

    for i, sponsor := range sponsors {
        if sponsor.Address == address {
            sponsors[i].Amount = sponsors[i].Amount.Add(amount)
            found = true
            break
        }
    }

    if !found {
        sponsors = append(sponsors, Sponsor{Address: address, Amount: amount})
    }

    totalDonated = totalDonated.Add(amount)

    totalDonations++

    sortSponsorsByAmount()

    if len(cities) > 0 {
        currentCityIndex++
        if currentCityIndex >= len(cities) {
            currentCityIndex = 0
        }
    }
}

type SponsorSlice []Sponsor

func (s SponsorSlice) Len() int {
	return len(s)
}

func (s SponsorSlice) Less(i, j int) bool {
	return s[i].Amount.AmountOf("ugnot") > s[j].Amount.AmountOf("ugnot")
}

func (s SponsorSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func sortSponsorsByAmount() {
	sort.Sort(SponsorSlice(sponsors))
}

func GetTopSponsors() []Sponsor {
	return sponsors
}

func CollectDonations() {
	config.AssertAuthorized()

	banker := std.GetBanker(std.BankerTypeRealmSend)

	ownerAddr := config.Address()
	banker.SendCoins(std.GetOrigPkgAddr(), ownerAddr, banker.GetCoins(std.GetOrigPkgAddr()))
}

func GetTotalDonations() std.Coins {
	return totalDonated
}

func GetDonationCount() int {
	return totalDonations
}

func Render(path string) string {
	out := ufmt.Sprintf("# Exploring %s!\n\n", cities[currentCityIndex].Name)

	out += renderAboutMe()
	out += "\n\n"
	out += renderTips()

	return out
}

func renderAboutMe() string {
	out := "<div class='rows-3'>"

	out += "<div style='position: relative; text-align: center;'>\n\n"

	out += ufmt.Sprintf("<div style='background-image: url(%s); background-size: cover; background-position: center; width: 100%%; height: 600px; position: relative; border-radius: 15px; overflow: hidden;'>\n\n", cities[currentCityIndex].URL)

	out += ufmt.Sprintf("<img src='%s' alt='my profile pic' style='width: 250px; height: auto; aspect-ratio: 1 / 1; object-fit: cover; border-radius: 50%%; border: 3px solid #1e1e1e; position: absolute; top: 75%%; left: 50%%; transform: translate(-50%%, -50%%);'>\n\n", pfp)

	out += "</div>\n\n"

	out += "<div>\n\n"
	out += aboutMe[0] + "\n\n"
	out += "</div>\n\n"

	out += "<div>\n\n"
	out += aboutMe[1] + "\n\n"
	out += "</div>\n\n"

	out += "</div><!-- /rows-3 -->\n\n"

	return out
}

func renderTips() string {
	out := `<div class="jumbotron" style="display: flex; flex-direction: column; justify-content: flex-start; align-items: center; padding-top: 40px; padding-bottom: 50px; text-align: center;">` + "\n\n"

	out += `<div class="rows-2" style="max-width: 500px; width: 100%; display: flex; flex-direction: column; justify-content: center; align-items: center;">` + "\n"

	out += `<h1 style="margin-bottom: 50px;">Help Me Travel The World</h1>` + "\n\n"

	out += renderTipsJar() + "\n"

    out += ufmt.Sprintf(`<strong style="font-size: 1.2em;">I am currently in %s, <br> tip the jar to send me somewhere else!</strong>`, cities[currentCityIndex].Name)
    
    out += `<br><span style="font-size: 1.2em; font-style: italic; margin-top: 10px; display: inline-block;">Click the jar, tip in GNOT coins, and watch my background change as I head to a new adventure!</span></p>` + "\n\n"
    
	out += renderSponsors()

	out += `</div><!-- /rows-2 -->` + "\n\n"

	out += `</div><!-- /jumbotron -->` + "\n"

	return out
}

func formatAddress(address string) string {
    if len(address) <= 8 {
        return address
    }
    return address[:4] + "..." + address[len(address)-4:]
}

func renderSponsors() string {
    out := `<h3 style="margin-top: 5px; margin-bottom: 20px">Sponsor Leaderboard</h3>` + "\n"

    if len(sponsors) == 0 {
        out += `<p style="text-align: center;">No sponsors yet. Be the first to tip the jar!</p>` + "\n"
    } else {
        numSponsors := len(sponsors)
        if numSponsors > maxSponsors {
            numSponsors = maxSponsors
        }

        out += `<ul style="list-style-type: none; padding: 0; border: 1px solid #ddd; border-radius: 8px; width: 100%; max-width: 300px; margin: 0 auto;">` + "\n"

        for i := 0; i < numSponsors; i++ {
            sponsor := sponsors[i]
            isLastItem := (i == numSponsors-1)

            padding := "10px 5px"
            border := "border-bottom: 1px solid #ddd;"

            if isLastItem {
                padding = "8px 5px"
                border = ""
            }

            out += ufmt.Sprintf(
                `<li style="padding: %s; %s text-align: left;">
                    <strong style="padding-left: 5px;">%d. %s</strong> 
                    <span style="float: right; padding-right: 5px;">%s</span>
                </li>`,
                padding, border, i+1, formatAddress(sponsor.Address.String()), sponsor.Amount.String(),
            )
        }

    }

    return out
}

func renderTipsJar() string {
	out := ufmt.Sprintf(`<a href="%s" target="_blank" style="display: block; text-decoration: none;">`, jarLink) + "\n"

	out += `<img src="https://i.ibb.co/4TH9zbw/tips-jar.png" alt="Tips Jar" style="width: 300px; height: auto; display: block; margin: 0 auto;">` + "\n"

	out += `</a>` + "\n"

	return out
}
