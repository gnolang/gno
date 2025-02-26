package home

import (
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/r/dragan/config"
)

type SocialLink struct {
	URL  string
	Text string
}

type Project struct {
	Name        string
	Description string
	URL         string
	ImageURL    string
}

var (
	aboutMe       string
	socialLinks   map[string]SocialLink
	gnoProjects   map[string]Project
	web3Projects  map[string]Project
	otherProjects map[string]Project
	textArt       string
	comingSoon    string
)

func init() {
	aboutMe = "I am a final-year student of Applied Software Engineering at the Faculty of Technical Sciences in Novi Sad, with a strong focus on blockchain and Web3 technologies.\n\n Since attending the Web3 Camp in Petnica, I have been actively exploring decentralized systems, smart contract development, and blockchain infrastructure.\n\n Additionally, I have experience in AI, particularly in NLP and neural networks. I have competed in and won multiple hackathons, including ETH Sofia, where I secured four sponsor bounties. Passionate about contributing to the blockchain ecosystem, I am eager to collaborate on innovative projects and continue expanding my expertise.\n\n"
	comingSoon = "![Coming Soon](https://img.freepik.com/free-vector/abstract-grunge-style-coming-soon-with-black-splatter_1017-26690.jpg)\n"

	socialLinks = map[string]SocialLink{
		"GitHub":   {URL: "https://github.com/Milosevic02", Text: "Explore my repositories and open-source contributions."},
		"LinkedIn": {URL: "https://www.linkedin.com/in/dragan-milosevic3/", Text: "Connect with me professionally."},
		"Email Me": {URL: "mailto:milosevicdragan002@gmail.com", Text: "Reach out for collaboration or inquiries."},
	}

	gnoProjects = make(map[string]Project)
	web3Projects = make(map[string]Project)
	otherProjects = make(map[string]Project)

	gnoProjects["Liberty Bridge"] = Project{
		Name:        "Liberty Bridge",
		Description: "Liberty Bridge was my first Web3 project, developed as part of the Web3 Bootcamp at Petnica. This project served as a centralized bridge between Ethereum and gno.land, enabling seamless asset transfers and fostering interoperability between the two ecosystems.\n\n The primary objective of Liberty Bridge was to address the challenges of connecting decentralized networks by implementing a user-friendly solution that simplified the process for users. The project incorporated mechanisms to securely transfer assets between the Ethereum and gno.land blockchains, ensuring efficiency and reliability while maintaining a centralized framework for governance and operations.\n\n Through this project, I gained hands-on knowledge of blockchain interoperability, Web3 protocols, and the intricacies of building solutions that bridge different blockchain ecosystems.\n\n",
		URL:         "https://gno.land",
		ImageURL:    "https://github.com/Milosevic02/LibertyBridge/raw/main/lb_banner.png",
	}

	gnoProjects["Liars Lie"] = Project{
		Name:        "Liars Lie",
		Description: "",
		URL:         "https://gno.land",
		ImageURL:    "https://img.freepik.com/free-vector/abstract-grunge-style-coming-soon-with-black-splatter_1017-26690.jpg\n",
	}

	web3Projects["Incognito"] = Project{
		Name:        "Incognito",
		Description: "Incognito is a Web3 platform built for Ethereum-based chains, designed to connect advertisers with users in a privacy-first and mutually beneficial way. Its modular architecture makes it easily expandable to other blockchains. Developed during the ETH Sofia Hackathon, it was recognized as a winning project for its innovation and impact.\n\nThe platform allows advertisers to send personalized ads while sharing a portion of the marketing budget with users. It uses machine learning to match users based on wallet activity, ensuring precise targeting. User emails are stored securely on-chain and never shared, prioritizing privacy and transparency.\n\nWith all campaign data stored on-chain, Incognito ensures decentralization and accountability. By rewarding users and empowering advertisers, it sets a new standard for fair and transparent blockchain-based advertising.\n\nDemo video of the project can be viewed at this link: [Incognito Demo Video](https://dorahacks.io/buidl/17725)",
		URL:         "https://github.com/Milosevic02/Incognito-ETHSofia",
		ImageURL:    "https://i.ibb.co/cS6RPW9L/Screenshot-2025-02-05-084304.png",
	}

	web3Projects["Solidity Exploits Research"] = Project{
		Name:        "Solidity Exploits Research",
		Description: "From August 1 to 10 2024, I had the incredible opportunity to attend the Web3 Camp in Petnica Science Center, where I worked on a research project focused on Solidity exploits. This project explored vulnerabilities in smart contracts, including reentrancy attacks, front-running, and integer overflows. The experience provided deep insights into blockchain security and best practices for writing secure Solidity code.\n\nYou can check out the presentation here: [Solidity Exploits Research](https://docs.google.com/presentation/d/1LziLmONP9oLpyRU5SrKUBCVqJ3yY1gOC/edit#slide=id.p1).",
		URL:         "https://docs.google.com/presentation/d/1LziLmONP9oLpyRU5SrKUBCVqJ3yY1gOC/edit#slide=id.p1",
		ImageURL:    "https://i.ibb.co/1t00VYMJ/Dragan-Milosevic-Solidity-Exploits-pptx.png", //Fix this
	}

	otherProjects["OtherProjects"] = Project{
		Name: "Highlight Projects",
		Description: "→ **BankBot - COMTRADE EDIT Hackathon - First Place** Smart banking chatbot built with Dialogflow, BERT, LangChain, and ChatGPT API. [Presentation](https://drive.google.com/file/d/11IDrka4hjhUCLfdF0EHRVgBx_YH-XSxi/view).\n\n" +
			"→ **Full-Stack Web Application Development**: Developed **Šped Helper** for a client, a full-stack app designed for a freight and shipping company, with business-critical functionalities.\n\n" +
			"→ **HardSkills Academy Hackathon - First Place**: Created a model for land cover classification using satellite images.\n\n" +
			"→ **BOSCH Future Mobility Challenge (Nov 2023 - Mar 2024)**: Led autonomous vehicle development, focusing on navigation and obstacle detection algorithms.\n\n" +
			"→ **Ethical Hacking Projects**: Developed ethical hacking tools, including a screenshot malware and ransomware simulation.\n\n" +
			"→ **Artificial Intelligence & Neural Networks**: Worked on various AI and NN projects as part of learning in the field.\n\n" +
			"🔍For more details on these projects, feel free to check my [LinkedIn](https://www.linkedin.com/in/dragan-milosevic3/).",
		URL:      "https://www.linkedin.com/in/dragan-milosevic3/",
		ImageURL: "",
	}
}

func Render(path string) string {
	var sb strings.Builder
	sb.WriteString("# Hi, I'm Dragan \n\n")
	sb.WriteString("---\n")
	sb.WriteString("## 👋 About me\n")
	sb.WriteString(aboutMe)
	sb.WriteString("---\n")
	sb.WriteString(renderProjects(gnoProjects, "⚙️ Gno Projects"))
	sb.WriteString("---\n")
	sb.WriteString(renderProjects(web3Projects, "🌐 Web3 Projects"))
	sb.WriteString("---\n")
	sb.WriteString(renderProjects(otherProjects, "📂 Other Projects"))
	sb.WriteString("---\n")
	sb.WriteString("## 🎮 Games\n")
	sb.WriteString("### " + "⏳" + comingSoon + "\n")
	sb.WriteString("---\n")
	sb.WriteString(renderRewardsAndAchievements())
	sb.WriteString(renderSocialLinks())

	return sb.String()
}

func renderRewardsAndAchievements() string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString("## 🏅 Rewards & Achievements\n\n")
	sb.WriteString("- 🏆 **ETH Sofia** – Won 4 sponsors’ bounties:\n")
	sb.WriteString("  - 🥇 1st place – Chainbase, iExec, SEEBLOCKS.eu\n")
	sb.WriteString("  - 🥉 3rd place – AI Open Track\n")
	sb.WriteString("- 🏆 **First place** in the COMTRADE EDIT competition for creating a bank chatbot with NLU capabilities.\n")
	sb.WriteString("- 🏆 **First place** in a hackathon on convolutional neural networks.\n")
	sb.WriteString("- 🚀 **Attended** Web3 Camp in Petnica.\n")
	sb.WriteString("- 🌟 **Participated** in the BOSCH FUTURE MOBILITY CHALLENGE.\n")
	sb.WriteString("---\n")

	return sb.String()
}

func renderSocialLinks() string {
	var sb strings.Builder
	sb.WriteString("## 🔗 Links\n\n")
	sb.WriteString("You can find me here:\n\n")
	sb.WriteString(ufmt.Sprintf("- 🐙 [GitHub](%s) - %s\n", socialLinks["GitHub"].URL, socialLinks["GitHub"].Text))
	sb.WriteString(ufmt.Sprintf("- 👜 [LinkedIn](%s) - %s\n", socialLinks["LinkedIn"].URL, socialLinks["LinkedIn"].Text))
	sb.WriteString(ufmt.Sprintf("- 📧 [Email Me](%s) - %s\n", socialLinks["Email Me"].URL, socialLinks["Email Me"].Text))
	sb.WriteString("\n")
	return sb.String()
}

func renderProjects(projectsMap map[string]Project, title string) string {
	var sb strings.Builder
	sb.WriteString(ufmt.Sprintf("## %s\n\n", title))
	for _, project := range projectsMap {
		sb.WriteString(ufmt.Sprintf("### [%s](%s)\n\n", project.Name, project.URL))
		if project.ImageURL != "" {
			sb.WriteString(ufmt.Sprintf("![%s](%s)\n\n", project.Name, project.ImageURL))
		}
		sb.WriteString(project.Description + "\n\n")
	}
	return sb.String()
}

func UpdateLink(name, newURL string) {
	if !isAuthorized(std.PrevRealm().Addr()) {
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

func UpdateAboutMe(text string) {
	if !isAuthorized(std.PrevRealm().Addr()) {
		panic(config.ErrUnauthorized)
	}

	aboutMe = text
}

func AddGnoProject(name, description, url, imageURL string) {
	if !isAuthorized(std.PrevRealm().Addr()) {
		panic(config.ErrUnauthorized)
	}
	project := Project{
		Name:        name,
		Description: description,
		URL:         url,
		ImageURL:    imageURL,
	}
	gnoProjects[name] = project
}

func DeleteGnoProject(projectName string) {
	if !isAuthorized(std.PrevRealm().Addr()) {
		panic(config.ErrUnauthorized)
	}

	if _, exists := gnoProjects[projectName]; !exists {
		panic("Project not found")
	}

	delete(gnoProjects, projectName)
}

func AddWeb3Project(name, description, url, imageURL string) {
	if !isAuthorized(std.PrevRealm().Addr()) {
		panic(config.ErrUnauthorized)
	}
	project := Project{
		Name:        name,
		Description: description,
		URL:         url,
		ImageURL:    imageURL,
	}
	web3Projects[name] = project
}

func RemoveWeb3Project(projectName string) {
	if !isAuthorized(std.PrevRealm().Addr()) {
		panic(config.ErrUnauthorized)
	}

	if _, exists := otherProjects[projectName]; !exists {
		panic("Project not found")
	}

	delete(web3Projects, projectName)
}

func AddOtherProject(name, description, url, imageURL string) {
	if !isAuthorized(std.PrevRealm().Addr()) {
		panic(config.ErrUnauthorized)
	}
	project := Project{
		Name:        name,
		Description: description,
		URL:         url,
		ImageURL:    imageURL,
	}
	otherProjects[name] = project
}

func RemoveOtherProject(projectName string) {
	if !isAuthorized(std.PrevRealm().Addr()) {
		panic(config.ErrUnauthorized)
	}

	if _, exists := otherProjects[projectName]; !exists {
		panic("Project not found")
	}

	delete(otherProjects, projectName)
}

func isAuthorized(addr std.Address) bool {

	return addr == config.Address() || addr == config.Backup()
}
