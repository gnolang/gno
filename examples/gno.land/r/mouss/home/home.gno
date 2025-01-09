package home

import (
	"bytes"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
	"gno.land/r/mouss/config"
)

type Profile struct {
	AboutMe   string
	Avatar    string
	Email     string
	Github    string
	LinkedIn  string
	Followers []std.Address
}

type Recipe struct {
	Name         string
	Origin       string
	Author       std.Address
	Ingredients  string
	Instructions string
	Tips         string
}

var (
	profile = Profile{
		AboutMe: "👋 I'm Mustapha, a contributor to gno.land project from France. " +
			"I'm passionate about coding, exploring new technologies, and contributing to open-source projects. " +
			"Besides my tech journey, I'm also a pizzaiolo 🍕 who loves cooking and savoring good food.",
		Avatar:    "https://github.com/mous1985/assets/blob/master/avatar.png?raw=true",
		Email:     "mustapha.benazzouz@outlook.fr",
		Github:    "https://github.com/mous1985",
		LinkedIn:  "https://www.linkedin.com/in/mustapha-benazzouz-88646887/",
		Followers: make([]std.Address, 0),
	}
	router  = mux.NewRouter()
	Recipes []*Recipe
)

const (
	RealmURL = "/r/mouss/home"
	Rec      = RealmURL + ":recipe/%d"
	//Followers = RealmURL + "$help&func=follow=%s"
)

func init() {
	router.HandleFunc("", renderHomepage)
	router.HandleFunc("recipe/{id}", renderRecipe)

}

func AddRecipe(name, origin string, author std.Address, ingredients, instructions string, tips string) string {
	if name == "" || len(ingredients) == 0 || len(instructions) == 0 {
		panic("Your	recipe is incomplet!")
	}
	recipe := Recipe{
		Name:         name,
		Origin:       origin,
		Author:       std.PrevRealm().Addr(),
		Ingredients:  ingredients,
		Instructions: instructions,
		Tips:         tips,
	}
	Recipes = append(Recipes, &recipe)
	return "Recipe added successfully!"

}

// func renderRecipes(res *mux.ResponseWriter, req *mux.Request) {
// 	var b bytes.Buffer
// 	writeNavigation(&b)

// 	b.WriteString("# Global Recipes Collection\n\n")
// 	b.WriteString("Share and discover recipes from around the world!\n\n")

// 	for _, recipe := range Recipes {
// 		b.WriteString(ufmt.Sprintf("## %s\n", recipe.Name))
// 		b.WriteString(ufmt.Sprintf("Origin: %s\n\n", recipe.Origin))
// 		b.WriteString(ufmt.Sprintf("Shared by: %s\n\n", recipe.Author))

// 		b.WriteString("### Ingredients\n")
// 		for _, ing := range recipe.Ingredients {
// 			b.WriteString("- " + ing + "\n")
// 		}
// 		b.WriteString("\n### Instructions\n")
// 		for i, inst := range recipe.Instructions {
// 			b.WriteString(ufmt.Sprintf("%d. %s\n", i+1, inst))
// 		}

// 		b.WriteString(ufmt.Sprintf("\nCooking Time: %s\n\n", recipe.CookingTime))
// 		if recipe.Tips != "" {
// 			b.WriteString(ufmt.Sprintf("💡 **Tips**: %s\n\n", recipe.Tips))
// 		}
// 		b.WriteString("------\n\n")
// 	}

// 	res.Write(b.String())
// }

func SetProfile(p Profile) {
	profile = p
}

func Follow(addr std.Address) error {
	if !isUser(std.PrevRealm().Addr()) {
		return config.ErrorUnauthorized
	}
	if contains(profile.Followers, addr) {
		return ufmt.Errorf("address %s is already following", addr)
	}
	profile.Followers = append(profile.Followers, addr)
	return nil
}

func isAuthorized(addr std.Address) bool {
	return addr == config.Address() || addr == config.Backup()
}

func isUser(addr std.Address) bool {
	return !isAuthorized(addr) && !contains(profile.Followers, addr)
}
func contains(slice []std.Address, item std.Address) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func renderRecipe(res *mux.ResponseWriter, req *mux.Request) {
	idStr := req.GetVar("id")
	id, _ := strconv.Atoi(idStr)
	if id >= len(Recipes) {
		res.Write("Invalid recipe ID")
		return
	}

	rec := Recipes[id]
	var b bytes.Buffer
	for _, recipe := range Recipes {

		b.WriteString("## " + recipe.Name + "\n")
		b.WriteString("**Author:** " + recipe.Author.String() + "\n")
		b.WriteString("### Origin: " + recipe.Origin + "\n")
		b.WriteString("### Ingredients: " + recipe.Ingredients + "\n")

		b.WriteString("\n### Instructions :" + recipe.Ingredients + "\n")

		if rec.Tips != "" {
			b.WriteString(ufmt.Sprintf("💡 **Tips**: %s\n\n", recipe.Tips))
		}
		b.WriteString("------\n\n")
	}

	res.Write(b.String())
}

func writeNavigation(b *bytes.Buffer) {
	b.WriteString(strings.Join([]string{
		"[Home]()",
		ufmt.Sprintf("[World kitchen](%s)", Rec),
		"[Hackerspace](https://github.com/gnolang/hackerspace/issues/86#issuecomment-2535795751)",
	}, " | ") + "\n\n------\n\n")
}

func writeGnoArt(b *bytes.Buffer) {
	b.WriteString("```\n")
	for _, line := range []string{
		"              -==++.                                                                  ",
		"             *@@@@=                                     @-                          -@",
		"            #@@@@@:       -==-.-- :-::===:   .-++-.     @-   .===:.- .-.-==-   .===:=@",
		"           #@@@@@@@:    -@@%**%@@ #@@#*#@@- *@@**@@*    @-  +%=::-*@ +@=-:-@* +%=::-*@",
		"          +@%#**#%@@    %@+   :@@ *@+   #@=+@%    %@+   @= :@:    -@ +%    +%.@:    -@",
		"          -:       -    *@%:..+@@ *@+   #@=-@@:  :@@=   @- .@=    =@ +@    *%.@=    =@",
		"          --:==+=-:=.    =%@%#*@@ *@+   #@+ =%@%%@%= #* %#=.:%*===*@ +%    +% -%*===*@",
		"          -++++=++++.    =-:::*@#  .     .    .::.   ..  ::   .::  .  .         .::  .",
		"           .-=+++=:     .*###%#=                                                      ",
		"              ::                                                                      ",
	} {
		b.WriteString(line + "\n")
	}
	b.WriteString("```\n------")
}
func renderHomepage(res *mux.ResponseWriter, req *mux.Request) {
	var b bytes.Buffer

	writeNavigation(&b)
	b.WriteString("# Welcome to my Homepage\n\n")
	b.WriteString("------\n\n")
	writeGnoArt(&b)
	b.WriteString("------\n\n")
	b.WriteString("## About Me\n------\n\n")
	b.WriteString(ufmt.Sprintf("![avatar](%s)\n\n", profile.Avatar))
	b.WriteString(profile.AboutMe + "\n\n------\n\n")

	b.WriteString("### Contact\n\n")
	b.WriteString(ufmt.Sprintf("- Email: %s\n", profile.Email))
	b.WriteString(ufmt.Sprintf("- GitHub: [@mous1985](%s)\n", profile.Github))
	b.WriteString(ufmt.Sprintf("- LinkedIn: [Mustapha](%s)\n", profile.LinkedIn))

	b.WriteString("------\n\n")
	b.WriteString(ufmt.Sprintf("👤 Followers: %d", len(profile.Followers)))

	res.Write(b.String())
}

func Render(path string) string {
	return router.Render(path)
}
