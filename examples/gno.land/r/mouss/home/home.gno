package home

import (
	"bytes"

	"gno.land/p/demo/mux"
)

var (
	router = mux.NewRouter()
)

const (
	RealmURL = "/r/mouss/home"
)

func init() {
	router.HandleFunc("", renderHomepage)
	router.HandleFunc("/abtMe", aboutme.renderAboutMe)

}

func renderHomepage(res *mux.ResponseWriter, req *mux.Request) {
	var b bytes.Buffer

	b.WriteString("[Home]() | [About Me](aboutMe) | [Hackerspace](https://github.com/gnolang/hackerspace/issues/86#issuecomment-2535795751)\n\n")
	b.WriteString("--------------------------------\n\n")
	b.WriteString("```\n")
	for _, line := range gnoArt() {
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("```\n")
	b.WriteString("--------------------------------")

	res.Write(b.String())
}

func gnoArt() []string {
	return []string{
		"    -==++.                                                                  ",
		"   *@@@@=                                     @-                          -@",
		"  #@@@@@:       -==-.-- :-::===:   .-++-.     @-   .===:.- .-.-==-   .===:=@",
		" #@@@@@@@:    -@@%**%@@ #@@#*#@@- *@@**@@*    @-  +%=::-*@ +@=-:-@* +%=::-*@",
		"+@%#**#%@@    %@+   :@@ *@+   #@=+@%    %@+   @= :@:    -@ +%    +%.@:    -@",
		"-:       -    *@%:..+@@ *@+   #@=-@@:  :@@=   @- .@=    =@ +@    *%.@=    =@",
		"--:==+=-:=.    =%@%#*@@ *@+   #@+ =%@%%@%= #* %#=.:%*===*@ +%    +% -%*===*@",
		"-++++=++++.    =-:::*@#  .     .    .::.   ..  ::   .::  .  .         .::  .",
		" .-=+++=:     .*###%#=                                                      ",
		"    ::                                                                      ",
	}
}

func Render(path string) string {
	return router.Render(path)
}
