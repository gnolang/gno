package home

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/ownable"
	"gno.land/p/moul/md"
	"gno.land/r/leon/hor"

	"gno.land/r/ursulovic/registry"
)

var (
	aboutMe       string
	selectedImage string
	Ownable       *ownable.Ownable

	githubUrl        string
	linkedinUrl      string
	imageUpdatePrice int64

	isValidUrl func(string) bool
)

func init() {
	Ownable = ownable.NewWithAddress(registry.MainAddress())

	aboutMe = "Hi, I'm Ivan Ursulovic, a computer engineering graduate, blockchain enthusiast, and backend developer specializing in ASP.NET. I love learning new things and taking on challenges."
	selectedImage = "https://i.ibb.co/W28NPkw/beograd.webp"

	githubUrl = "https://github.com/ursulovic"
	linkedinUrl = "https://www.linkedin.com/in/ivan-ursulovic-953310190/"
	imageUpdatePrice = 5000000
	isValidUrl = defaultURLValidation
	hor.Register("Ivan's Home Realm", "Welcome to my Home Realm!")
}

func Render(s string) string {
	var sb strings.Builder
	sb.WriteString(renderAboutMe())
	sb.WriteString(renderSelectedImage())
	sb.WriteString(renderContactsUrl())
	return sb.String()
}

func defaultURLValidation(url string) bool {
	const urlPrefix string = "https://i.ibb.co/"

	if !strings.HasPrefix(url, urlPrefix) {
		return false
	}

	if !(strings.HasSuffix(url, ".jpg") ||
		strings.HasSuffix(url, ".png") ||
		strings.HasSuffix(url, ".gif") ||
		strings.HasSuffix(url, ".webp")) {
		return false
	}

	urlPath := strings.TrimPrefix(url, "https://i.ibb.co/")
	parts := strings.Split(urlPath, "/")

	if len(parts) != 2 || len(parts[0]) == 0 || len(parts[1]) == 0 {
		return false
	}

	return true
}

func UpdateSelectedImage(url string) {
	if !isValidUrl(url) {
		panic("Url is not valid!")
	}

	sentCoins := std.OriginSend()

	if len(sentCoins) != 1 || sentCoins.AmountOf("ugnot") != imageUpdatePrice {
		panic("Please send exactly " + strconv.Itoa(int(imageUpdatePrice)) + " ugnot")
	}

	selectedImage = url
}

func renderSelectedImage() string {
	var sb strings.Builder

	sb.WriteString(md.HorizontalRule())
	sb.WriteString("\n")

	sb.WriteString(md.H2("📸 Featured Image"))
	sb.WriteString("\n")

	sb.WriteString(md.Image("", selectedImage))
	sb.WriteString("\n")

	sb.WriteString(md.H4("✨ " + md.Link("Change this image for "+strconv.Itoa(int(imageUpdatePrice/1000000))+" GNOT. To update, set a direct image URL from ImgBB.", "https://gno.studio/connect/view/gno.land/r/ursulovic/home?network=portal-loop") + " ✨"))

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
	Ownable.AssertOwned()
	githubUrl = url
}

func UpdateLinkedinUrl(url string) {
	Ownable.AssertOwned()
	linkedinUrl = url
}

func UpdateAboutMe(text string) {
	Ownable.AssertOwned()
	aboutMe = text
}

func UpdateImagePrice(newPrice int64) {
	Ownable.AssertOwned()
	imageUpdatePrice = newPrice
}

func UpdateIsValidUrlFunction(f func(string) bool) {
	Ownable.AssertOwned()
	isValidUrl = f
}
