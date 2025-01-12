package home

import (
	"std"
	"strings"
	"strconv"

	"gno.land/p/moul/md"
	"gno.land/r/leon/hof"
	"gno.land/p/demo/ownable"

	"gno.land/r/ursulovic/registry"
)

var (
	aboutMe				string
	selectedPicture		string
	owner 				*ownable.Ownable

	githubUrl			string
	linkedinUrl			string
	imageUpdatePrice int64 = 1000000
)

func init() {
	owner = ownable.NewWithAddress(registry.MainAddress())

	aboutMe = "Hi, I'm Ivan Ursulovic, a computer engineering graduate, blockchain enthusiast, and backend developer specializing in ASP.NET. I love learning new things and taking on challenges."
	selectedPicture = "https://i.ibb.co/W28NPkw/beograd.webp"

	githubUrl = "https://github.com/ursulovic"
	linkedinUrl = "https://www.linkedin.com/in/ivan-ursulovic-953310190/"

	hof.Register()
}

func Render(s string) string {
	var sb strings.Builder
	sb.WriteString(renderAboutMe())
	sb.WriteString(renderSelectedImage())
	sb.WriteString(renderContactsUrl())
	return sb.String()
}

func isValidURL(url string) bool {
    if !strings.HasPrefix(url, "https://i.ibb.co/") {
        return false
    }

    if !(strings.HasSuffix(url, ".jpg") || 
         strings.HasSuffix(url, ".png") || 
         strings.HasSuffix(url, ".gif") ||
         strings.HasSuffix(url, ".webp")) { // Add webp since your current image uses it
        return false
    }

    urlPath := strings.TrimPrefix(url, "https://i.ibb.co/")
    parts := strings.Split(urlPath, "/")

    if len(parts) != 2 || len(parts[0]) == 0 || len(parts[1]) == 0 {
        return false
    }

    return true
}

func UpdateSelectedPicture(url string) {
	if !isValidURL(url) {
		panic("Url is not valid!")
	}

	sentCoins := std.GetOrigSend()
	
	if len(sentCoins) != 1 || sentCoins[0].Denom != "ugnot" || sentCoins[0].Amount != imageUpdatePrice {
		panic("Please send exactly " + strconv.Itoa(int(imageUpdatePrice)) + " ugnot")
	}

	selectedPicture = url
}

func renderSelectedImage() string {
	var sb strings.Builder
	
	sb.WriteString(md.HorizontalRule())
	sb.WriteString("\n")
	
	sb.WriteString(md.H2("📸 Featured Image"))
	sb.WriteString("\n")
	
	sb.WriteString(md.Image("Belgrade Skyline", selectedPicture))
	sb.WriteString("\n")
	
	sb.WriteString(md.H4("✨ " + md.Link("Change this picture for 1 GNOT", "https://www.google.com") + " ✨"))
	
	return sb.String()
}

func renderAboutMe() string {
	var sb strings.Builder
	
	sb.WriteString(md.H1("👋 Welcome to Ivan's Homepage!"))
	sb.WriteString("\n")
	
	sb.WriteString(md.H2("👨‍💻 About Me"))
	sb.WriteString("\n")
	
	sb.WriteString(md.Blockquote(aboutMe))
	
	return sb.String()
}

func renderContactsUrl() string {
	var sb strings.Builder
	
	sb.WriteString(md.HorizontalRule())
	sb.WriteString("\n")
	
	sb.WriteString(md.H2("🔗 Let's Connect"))
	sb.WriteString("\n")
	
	items := []string{
		"🐙 " + md.Link("GitHub", githubUrl),
		"💼 " + md.Link("LinkedIn", linkedinUrl),
	}
	sb.WriteString(md.BulletList(items))
	
	return sb.String()
}

func UpdateGithubUrl(url string) {
	owner.AssertCallerIsOwner()
	githubUrl = url
}

func UpdateLinkedinUrl(url string) {
	owner.AssertCallerIsOwner()
	linkedinUrl = url
}

func UpdateAboutMe(text string) {
	owner.AssertCallerIsOwner()
	aboutMe = text
}

func UpdateImagePrice(newPrice int64) {
	owner.AssertCallerIsOwner()
	imageUpdatePrice = newPrice
}