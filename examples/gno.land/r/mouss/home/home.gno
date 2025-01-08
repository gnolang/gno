package home

import (
	"bytes"
	"std"

	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
	"gno.land/r/mouss/config"
)

var (
	abtMe     string
	avatar    string // profile picture
	followers []std.Address

	router = mux.NewRouter()
)

const (
	RealmURL  = "/r/mouss/home"
	AbtMe     = RealmURL + ":abtMe/"
	Followers = RealmURL + "$help&func=follow=%s"
)

func init() {
	router.HandleFunc("", renderHomepage)
	router.HandleFunc("abtMe/", renderAboutMe)
	abtMe = "👋 I'm Mustapha, a contributor to gno.land project from France. I’m passionate about coding, exploring new technologies, and contributing to open-source projects. Besides my tech journey, I’m also a pizzaiolo 🍕 who loves cooking and savoring good food."
	avatar = "https://github.com/mous1985/assets/blob/master/avatar.png?raw=true"

}
func Follow(addr std.Address) {

	if !isUser(std.PrevRealm().Addr()) {
		panic(config.ErrorUnauthorized)
	}
	followers = append(followers, addr)

}
func isAuthorized(addr std.Address) bool {
	return addr == config.Address() || addr == config.Backup()
}
func isUser(addr std.Address) bool {
	return !isAuthorized(addr) && !contains(followers, addr)
}

func contains(slice []std.Address, item std.Address) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
func renderHomepage(res *mux.ResponseWriter, req *mux.Request) {
	var b bytes.Buffer
	link := ufmt.Sprintf(AbtMe)
	b.WriteString("[Home]() | " + ufmt.Sprintf("[About Me](%s)", link) + " | [Hackerspace](https://github.com/gnolang/hackerspace/issues/86#issuecomment-2535795751)\n\n")
	b.WriteString("------\n\n")
	b.WriteString("# Welcome to my Homepage\n\n")
	b.WriteString("------\n\n")
	b.WriteString("```\n")
	for _, line := range gnoArt() {
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("```\n")
	b.WriteString("------")

	res.Write(b.String())
}
func renderAboutMe(res *mux.ResponseWriter, req *mux.Request) {
	var b bytes.Buffer
	b.WriteString("## About Me\n")
	b.WriteString("------\n\n")
	b.WriteString("![avatar](" + avatar + ")\n\n")
	b.WriteString(abtMe + "\n\n")
	b.WriteString("------\n\n")
	b.WriteString("### Contact\n\n")
	b.WriteString("- e-mail:mustapha.benazzouz@outlook.fr\n")
	b.WriteString("- github : [@mous1985](https://github.com/mous1985)\n")
	b.WriteString("- LinkedIn : [Mustapha](https://www.linkedin.com/in/mustapha-benazzouz-88646887/)\n")
	b.WriteString("------\n\n")
	b.WriteString("👤followers: " + ufmt.Sprintf("%d", len(followers)))

	res.Write(b.String())
}

func gnoArt() []string {
	return []string{

		"                -==++.                                                                  ",
		"               *@@@@=                                     @-                          -@",
		"              #@@@@@:       -==-.-- :-::===:   .-++-.     @-   .===:.- .-.-==-   .===:=@",
		"             #@@@@@@@:    -@@%**%@@ #@@#*#@@- *@@**@@*    @-  +%=::-*@ +@=-:-@* +%=::-*@",
		"            +@%#**#%@@    %@+   :@@ *@+   #@=+@%    %@+   @= :@:    -@ +%    +%.@:    -@",
		"            -:       -    *@%:..+@@ *@+   #@=-@@:  :@@=   @- .@=    =@ +@    *%.@=    =@",
		"            --:==+=-:=.    =%@%#*@@ *@+   #@+ =%@%%@%= #* %#=.:%*===*@ +%    +% -%*===*@",
		"            -++++=++++.    =-:::*@#  .     .    .::.   ..  ::   .::  .  .         .::  .",
		"             .-=+++=:     .*###%#=                                                      ",
		"                ::                                                                      ",
	}

}

func Render(path string) string {
	return router.Render(path)
}
