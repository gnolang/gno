package home

import (
	"std"
	"strings"

	"gno.land/p/moul/md"
	"gno.land/r/leon/hof"
	"gno.land/p/demo/ownable"

	"gno.land/r/ursulovic/registry"
)

var (
	phrases    [10]string
	aboutMe    string
	pfp        string
	pfpCaption string
	owner *ownable.Ownable
)

func init() {
	owner = ownable.NewWithAddress(registry.MainAddress())

	phrases[0] = "Why did the blockchain go to therapy? It had a lot of unprocessed transactions."
	phrases[1] = "Why don’t blockchains ever tell secrets? They use smart contracts – always open and never closed!"
	phrases[2] = "In Gnoland, we don’t worry about bugs – they’re just features waiting to be integrated!"
	phrases[3] = "Why did the smart contract break up with traditional finance? It was a long-term contract – no longer sustainable."
	phrases[4] = "Why do blockchains make great comedians? They always deliver the punchline at just the right time!"
	phrases[5] = "Why did the blockchain apply for a job? It wanted to become the most reliable ledger around!"
	phrases[6] = "Why did the blockchain file a police report? It had a case of identity theft – someone stole its private keys!"
	phrases[7] = "In Gnoland, our smart contracts aren’t just written in code – they’re also written in poetry!"
	phrases[8] = "Why did the blockchain join a gym? To get a better hash rate!"
	phrases[9] = "Why did the blockchain break up with traditional finance? It was looking for a more transparent relationship."

	aboutMe = "Hello everyone, my name is Ivan Ursulovic. I am a Computer Engineering graduate from the School of Computing at Union University. I have experience working as a backend ASP.NET developer, but beyond that, I am a big Web3 enthusiast!"
	pfp = "https://i.ibb.co/W28NPkw/beograd.webp"
	pfpCaption = "Belgrade, the city I come from"
	hof.Register()
}

//Update
func UpdatePfp(url string, caption string) {
	owner.AssertCallerIsOwner()
	pfp = url
	pfpCaption = caption
}

func UpdateAboutMe(abtMe string) {
	owner.AssertCallerIsOwner()
	aboutMe = abtMe
}

func UpdatePhrase(index int, newPhrase string) {
	owner.AssertCallerIsOwner()
	if index < 0 || index >= len(phrases) {
		panic("index out of boundaries")
	}
	phrases[index] = newPhrase
}


//Rendering
func Render(path string) string {
	var out strings.Builder
	out.WriteString(renderAboutMe())
	out.WriteString(renderPhrase())
	out.WriteString(md.Image(pfpCaption, pfp))
	out.WriteString(md.Italic("Belgrade, the city I come from"))
	return out.String()
}

func renderAboutMe() string {
	var out strings.Builder
	out.WriteString(md.H1(`Ivan's Homepage`))
	out.WriteString(md.H4(aboutMe))

	return out.String()
}

func renderPhrase() string {
	var out strings.Builder
	phrase := phrases[int(std.GetHeight())%len(phrases)]

	out.WriteString(md.H3(`Randomly selected blockchain wisdom:`) + md.Bold(phrase))

	return out.String()
}
