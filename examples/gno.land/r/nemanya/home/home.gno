package home

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"strings"

	"gno.land/p/nt/ufmt"
	"gno.land/r/nemanya/config"
)

type SocialLink struct {
	URL  string
	Text string
}

type Sponsor struct {
	Address address
	Amount  chain.Coins
}

type Project struct {
	Name        string
	Description string
	URL         string
	ImageURL    string
	Sponsors    map[address]Sponsor
}

var (
	textArt        string
	aboutMe        string
	sponsorInfo    string
	socialLinks    map[string]SocialLink
	gnoProjects    map[string]Project
	otherProjects  map[string]Project
	totalDonations chain.Coins
)

func init() {
	textArt = renderTextArt()
	aboutMe = "I am a student of IT at Faculty of Sciences in Novi Sad, Serbia. My background is mainly in web and low-level programming, but since Web3 Bootcamp at Petnica this year I've been actively learning about blockchain and adjacent technologies. I am excited about contributing to the gno.land ecosystem and learning from the community.\n\n"
	sponsorInfo = "You can sponsor a project by sending GNOT to this address. Your sponsorship will be displayed on the project page. Thank you for supporting the development of gno.land!\n\n"

	socialLinks = map[string]SocialLink{
		"GitHub":   {URL: "https://github.com/Nemanya8", Text: "Explore my repositories and open-source contributions."},
		"LinkedIn": {URL: "https://www.linkedin.com/in/nemanjamatic/", Text: "Connect with me professionally."},
		"Email Me": {URL: "mailto:matic.nemanya@gmail.com", Text: "Reach out for collaboration or inquiries."},
	}

	gnoProjects = make(map[string]Project)
	otherProjects = make(map[string]Project)

	gnoProjects["Liberty Bridge"] = Project{
		Name:        "Liberty Bridge",
		Description: "Liberty Bridge was my first Web3 project, developed as part of the Web3 Bootcamp at Petnica. This project served as a centralized bridge between Ethereum and gno.land, enabling seamless asset transfers and fostering interoperability between the two ecosystems.\n\n The primary objective of Liberty Bridge was to address the challenges of connecting decentralized networks by implementing a user-friendly solution that simplified the process for users. The project incorporated mechanisms to securely transfer assets between the Ethereum and gno.land blockchains, ensuring efficiency and reliability while maintaining a centralized framework for governance and operations.\n\n Through this project, I gained hands-on knowledge of blockchain interoperability, Web3 protocols, and the intricacies of building solutions that bridge different blockchain ecosystems.\n\n",
		URL:         "https://gno.land",
		ImageURL:    "https://github.com/Milosevic02/LibertyBridge/raw/main/lb_banner.png",
		Sponsors:    make(map[address]Sponsor),
	}

	otherProjects["Incognito"] = Project{
		Name:        "Incognito",
		Description: "Incognito is a Web3 platform built for Ethereum-based chains, designed to connect advertisers with users in a privacy-first and mutually beneficial way. Its modular architecture makes it easily expandable to other blockchains. Developed during the ETH Sofia Hackathon, it was recognized as a winning project for its innovation and impact.\n\n The platform allows advertisers to send personalized ads while sharing a portion of the marketing budget with users. It uses machine learning to match users based on wallet activity, ensuring precise targeting. User emails are stored securely on-chain and never shared, prioritizing privacy and transparency.\n\n With all campaign data stored on-chain, Incognito ensures decentralization and accountability. By rewarding users and empowering advertisers, it sets a new standard for fair and transparent blockchain-based advertising.",
		URL:         "https://github.com/Milosevic02/Incognito-ETHSofia",
		ImageURL:    "",
		Sponsors:    make(map[address]Sponsor),
	}
}

func Render(_ string) string {
	var sb strings.Builder
	sb.WriteString("# Hi, I'm\n")
	sb.WriteString(textArt)
	sb.WriteString("---\n")
	sb.WriteString("## About me\n")
	sb.WriteString(aboutMe)
	sb.WriteString(sponsorInfo)
	sb.WriteString(ufmt.Sprintf("# Total Sponsor Donations: %s\n", totalDonations.String()))
	sb.WriteString("---\n")
	sb.WriteString(renderProjects(gnoProjects, "Gno Projects"))
	sb.WriteString("---\n")
	sb.WriteString(renderProjects(otherProjects, "Other Projects"))
	sb.WriteString("---\n")
	sb.WriteString(renderSocialLinks())

	return sb.String()
}

func renderTextArt() string {
	var sb strings.Builder
	sb.WriteString("```\n")
	sb.WriteString("      ___           ___           ___           ___           ___           ___           ___     \n")
	sb.WriteString("     /\\__\\         /\\  \\         /\\__\\         /\\  \\         /\\__\\         |\\__\\         /\\  \\    \n")
	sb.WriteString("    /::|  |       /::\\  \\       /::|  |       /::\\  \\       /::|  |        |:|  |       /::\\  \\   \n")
	sb.WriteString("   /:|:|  |      /:/\\:\\  \\     /:|:|  |      /:/\\:\\  \\     /:|:|  |        |:|  |      /:/\\:\\  \\  \n")
	sb.WriteString("  /:/|:|  |__   /::\\~\\:\\  \\   /:/|:|__|__   /::\\~\\:\\  \\   /:/|:|  |__      |:|__|__   /::\\~\\:\\  \\ \n")
	sb.WriteString(" /:/ |:| /\\__\\ /:/\\:\\ \\:\\__\\ /:/ |::::\\__\\ /:/\\:\\ \\:\\__\\ /:/ |:| /\\__\\     /::::\\__\\ /:/\\:\\ \\:\\__\\\n")
	sb.WriteString(" \\/__|:|/:/  / \\:\\~\\:\\ \\/__/ \\/__/~~/:/  / \\/__\\:\\/:/  / \\/__|:|/:/  /    /:/~~/~    \\/__\\:\\/:/  / \n")
	sb.WriteString("     |:/:/  /   \\:\\ \\:\\__\\         /:/  /       \\::/  /      |:/:/  /    /:/  /           \\::/  /  \n")
	sb.WriteString("     |::/  /     \\:\\ \\/__/        /:/  /        /:/  /       |::/  /     \\/__/            /:/  /   \n")
	sb.WriteString("     /:/  /       \\:\\__\\         /:/  /        /:/  /        /:/  /                      /:/  /    \n")
	sb.WriteString("     \\/__/         \\/__/         \\/__/         \\/__/         \\/__/                       \\/__/     \n")
	sb.WriteString("\n```\n")
	return sb.String()
}

func renderSocialLinks() string {
	var sb strings.Builder
	sb.WriteString("## Links\n\n")
	sb.WriteString("You can find me here:\n\n")
	sb.WriteString(ufmt.Sprintf("- [GitHub](%s) - %s\n", socialLinks["GitHub"].URL, socialLinks["GitHub"].Text))
	sb.WriteString(ufmt.Sprintf("- [LinkedIn](%s) - %s\n", socialLinks["LinkedIn"].URL, socialLinks["LinkedIn"].Text))
	sb.WriteString(ufmt.Sprintf("- [Email Me](%s) - %s\n", socialLinks["Email Me"].URL, socialLinks["Email Me"].Text))
	sb.WriteString("\n")
	return sb.String()
}

func renderProjects(projectsMap map[string]Project, title string) string {
	var sb strings.Builder
	sb.WriteString(ufmt.Sprintf("## %s\n\n", title))
	for _, project := range projectsMap {
		if project.ImageURL != "" {
			sb.WriteString(ufmt.Sprintf("![%s](%s)\n\n", project.Name, project.ImageURL))
		}
		sb.WriteString(ufmt.Sprintf("### [%s](%s)\n\n", project.Name, project.URL))
		sb.WriteString(project.Description + "\n\n")

		if len(project.Sponsors) > 0 {
			sb.WriteString(ufmt.Sprintf("#### %s Sponsors\n", project.Name))
			for _, sponsor := range project.Sponsors {
				sb.WriteString(ufmt.Sprintf("- %s: %s\n", sponsor.Address.String(), sponsor.Amount.String()))
			}
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func UpdateLink(_ realm, name, newURL string) {
	if !isAuthorized(runtime.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}

	if _, exists := socialLinks[name]; !exists {
		panic("Link with the given name does not exist")
	}

	socialLinks[name] = SocialLink{
		URL:  newURL,
		Text: socialLinks[name].Text,
	}
}

func UpdateAboutMe(_ realm, text string) {
	if !isAuthorized(runtime.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}

	aboutMe = text
}

func AddGnoProject(_ realm, name, description, url, imageURL string) {
	if !isAuthorized(runtime.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}
	project := Project{
		Name:        name,
		Description: description,
		URL:         url,
		ImageURL:    imageURL,
		Sponsors:    make(map[address]Sponsor),
	}
	gnoProjects[name] = project
}

func DeleteGnoProject(_ realm, projectName string) {
	if !isAuthorized(runtime.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}

	if _, exists := gnoProjects[projectName]; !exists {
		panic("Project not found")
	}

	delete(gnoProjects, projectName)
}

func AddOtherProject(_ realm, name, description, url, imageURL string) {
	if !isAuthorized(runtime.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}
	project := Project{
		Name:        name,
		Description: description,
		URL:         url,
		ImageURL:    imageURL,
		Sponsors:    make(map[address]Sponsor),
	}
	otherProjects[name] = project
}

func RemoveOtherProject(_ realm, projectName string) {
	if !isAuthorized(runtime.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}

	if _, exists := otherProjects[projectName]; !exists {
		panic("Project not found")
	}

	delete(otherProjects, projectName)
}

func isAuthorized(addr address) bool {
	return addr == config.Address() || addr == config.Backup()
}

func SponsorGnoProject(_ realm, projectName string) {
	address_XXX := runtime.OriginCaller()
	amount := banker.OriginSend()

	if amount.AmountOf("ugnot") == 0 {
		panic("Donation must include GNOT")
	}

	project, exists := gnoProjects[projectName]
	if !exists {
		panic("Gno project not found")
	}

	project.Sponsors[address_XXX] = Sponsor{
		Address: address_XXX,
		Amount:  project.Sponsors[address_XXX].Amount.Add(amount),
	}

	totalDonations = totalDonations.Add(amount)

	gnoProjects[projectName] = project
}

func SponsorOtherProject(_ realm, projectName string) {
	address_XXX := runtime.OriginCaller()
	amount := banker.OriginSend()

	if amount.AmountOf("ugnot") == 0 {
		panic("Donation must include GNOT")
	}

	project, exists := otherProjects[projectName]
	if !exists {
		panic("Other project not found")
	}

	project.Sponsors[address_XXX] = Sponsor{
		Address: address_XXX,
		Amount:  project.Sponsors[address_XXX].Amount.Add(amount),
	}

	totalDonations = totalDonations.Add(amount)

	otherProjects[projectName] = project
}

func Withdraw(_ realm) string {
	realmAddress := runtime.PreviousRealm().Address()
	if !isAuthorized(realmAddress) {
		panic(config.ErrUnauthorized)
	}

	banker_ := banker.NewBanker(banker.BankerTypeRealmSend)
	coins := banker_.GetCoins(realmAddress)

	if len(coins) == 0 {
		return "No coins available to withdraw"
	}

	banker_.SendCoins(realmAddress, config.Address(), coins)

	return "Successfully withdrew all coins to config address"
}
