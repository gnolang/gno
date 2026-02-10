package home

import (
	"sort"
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
	"gno.land/r/leon/hor"
	"gno.land/r/sys/users"

	"gno.land/r/stefann/registry"
)

type City struct {
	Name string
	URL  string
}

type Sponsor struct {
	Address address
	Amount  std.Coins
}

type Profile struct {
	aboutMe []string
}

type Travel struct {
	cities           []City
	currentCityIndex int
	jarLink          string
}

type Sponsorship struct {
	maxSponsors    int
	sponsors       *avl.Tree
	DonationsCount int
	sponsorsCount  int
}

var (
	profile     Profile
	travel      Travel
	sponsorship Sponsorship
	owner       *ownable.Ownable
)

func init() {
	owner = ownable.NewWithAddress(registry.MainAddr())
	hor.Register("Stefann's Home Realm", "")

	profile = Profile{
		aboutMe: []string{
			`## About Me`,
			`### Hey there! I’m Stefan, a student of Computer Science. I’m all about exploring and adventure — whether it’s diving into the latest tech or discovering a new city, I’m always up for the challenge!`,

			`## Contributions`,
			`### I'm just getting started, but you can follow my journey through gno.land right [here](https://github.com/gnolang/hackerspace/issues/94) 🔗`,
		},
	}

	travel = Travel{
		cities: []City{
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
		},
		currentCityIndex: 0,
		jarLink:          "https://TODO", // This value should be injected through UpdateJarLink after deployment.
	}

	sponsorship = Sponsorship{
		maxSponsors:    3,
		sponsors:       avl.NewTree(),
		DonationsCount: 0,
		sponsorsCount:  0,
	}
}

func UpdateCities(newCities []City) {
	owner.AssertOwnedByPrevious()
	travel.cities = newCities
}

func AddCities(newCities ...City) {
	owner.AssertOwnedByPrevious()

	travel.cities = append(travel.cities, newCities...)
}

func UpdateJarLink(newLink string) {
	owner.AssertOwnedByPrevious()
	travel.jarLink = newLink
}

func UpdateAboutMe(aboutMeStr string) {
	owner.AssertOwnedByPrevious()
	profile.aboutMe = strings.Split(aboutMeStr, "|")
}

func AddAboutMeRows(newRows ...string) {
	owner.AssertOwnedByPrevious()

	profile.aboutMe = append(profile.aboutMe, newRows...)
}

func UpdateMaxSponsors(newMax int) {
	owner.AssertOwnedByPrevious()
	if newMax <= 0 {
		panic("maxSponsors must be greater than zero")
	}
	sponsorship.maxSponsors = newMax
}

func Donate() {
	address := std.OriginCaller()
	amount := std.OriginSend()

	if amount.AmountOf("ugnot") == 0 {
		panic("Donation must include GNOT")
	}

	existingAmount, exists := sponsorship.sponsors.Get(address.String())
	if exists {
		updatedAmount := existingAmount.(std.Coins).Add(amount)
		sponsorship.sponsors.Set(address.String(), updatedAmount)
	} else {
		sponsorship.sponsors.Set(address.String(), amount)
		sponsorship.sponsorsCount++
	}

	travel.currentCityIndex++
	sponsorship.DonationsCount++

	banker := std.NewBanker(std.BankerTypeRealmSend)
	ownerAddr := registry.MainAddr()
	banker.SendCoins(std.CurrentRealm().Address(), ownerAddr, banker.GetCoins(std.CurrentRealm().Address()))
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

func GetTopSponsors() []Sponsor {
	var sponsorSlice SponsorSlice

	sponsorship.sponsors.Iterate("", "", func(key string, value any) bool {
		addr := address(key)
		amount := value.(std.Coins)
		sponsorSlice = append(sponsorSlice, Sponsor{Address: addr, Amount: amount})
		return false
	})

	sort.Sort(sponsorSlice)
	return sponsorSlice
}

func GetTotalDonations() int {
	total := 0
	sponsorship.sponsors.Iterate("", "", func(key string, value any) bool {
		total += int(value.(std.Coins).AmountOf("ugnot"))
		return false
	})
	return total
}

func Render(path string) string {
	out := ufmt.Sprintf("# Exploring %s!\n\n", travel.cities[travel.currentCityIndex].Name)

	out += renderAboutMe()
	out += "\n\n"
	out += renderTips()

	return out
}

func renderAboutMe() string {
	out := ""

	out += ufmt.Sprintf("![Current Location](%s)\n\n", travel.cities[travel.currentCityIndex%len(travel.cities)].URL)

	for _, rows := range profile.aboutMe {
		out += rows + "\n\n"
	}

	return out
}

func renderTips() string {
	out := "# Help Me Travel The World\n\n"

	out += ufmt.Sprintf("## I am currently in %s, tip the jar to send me somewhere else!\n\n", travel.cities[travel.currentCityIndex].Name)
	out += "### **Click** the jar, **tip** in GNOT coins, and **watch** my background change as I head to a new adventure!\n\n"

	out += renderTipsJar() + "\n\n"

	out += renderSponsors()

	return out
}

func formatAddress(address string) string {
	if len(address) <= 8 {
		return address
	}
	return address[:4] + "..." + address[len(address)-4:]
}

func getDisplayName(addr address) string {
	if user := users.ResolveAddress(addr); user != nil {
		return user.Name()
	}
	return formatAddress(addr.String())
}

func formatAmount(amount std.Coins) string {
	ugnot := amount.AmountOf("ugnot")
	if ugnot >= 1000000 {
		gnot := float64(ugnot) / 1000000
		return ufmt.Sprintf("`%v`*GNOT*", gnot)
	}
	return ufmt.Sprintf("`%d`*ugnot*", ugnot)
}

func renderSponsors() string {
	out := "## Sponsor Leaderboard\n\n"

	if sponsorship.sponsorsCount == 0 {
		return out + "No sponsors yet. Be the first to tip the jar!\n"
	}

	topSponsors := GetTopSponsors()
	numSponsors := len(topSponsors)
	if numSponsors > sponsorship.maxSponsors {
		numSponsors = sponsorship.maxSponsors
	}

	for i := 0; i < numSponsors; i++ {
		sponsor := topSponsors[i]
		position := ""
		switch i {
		case 0:
			position = "🥇"
		case 1:
			position = "🥈"
		case 2:
			position = "🥉"
		default:
			position = ufmt.Sprintf("%d.", i+1)
		}

		out += ufmt.Sprintf("%s **%s** - %s\n\n",
			position,
			getDisplayName(sponsor.Address),
			formatAmount(sponsor.Amount),
		)
	}

	return out + "\n"
}

func renderTipsJar() string {
	return ufmt.Sprintf("[![Tips Jar](https://i.ibb.co/4TH9zbw/tips-jar.png)](%s)", travel.jarLink)
}
