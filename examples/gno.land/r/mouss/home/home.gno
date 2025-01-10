package home

import (
	"std"
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

const (
	RealmURL = "/r/mouss/home"
	Rec      = RealmURL + ":recipe/"
)

var (
	profile = Profile{
		AboutMe:   "👋 I'm Mustapha, a contributor to gno.land project from France. I'm passionate about coding, exploring new technologies, and contributing to open-source projects. Besides my tech journey, I'm also a pizzaiolo 🍕 who loves cooking and savoring good food.",
		Avatar:    "https://github.com/mous1985/assets/blob/master/avatar.png?raw=true",
		Email:     "mustapha.benazzouz@outlook.fr",
		Github:    "https://github.com/mous1985",
		LinkedIn:  "https://www.linkedin.com/in/mustapha-benazzouz-88646887/",
		Followers: make([]std.Address, 0),
	}
	router          = mux.NewRouter()
	recipes         []*Recipe
	margheritaPizza = &Recipe{
		Name:         "Authentic Margherita Pizza 🤌",
		Origin:       "Naples, 🇮🇹",
		Author:       config.Address(),
		Ingredients:  " \n\n- 1kg 00 flour\n\n- 500ml water\n\n- 3g fresh yeast\n\n- 20g sea salt\n\n- San Marzano tomatoes\n\n- Fresh buffalo mozzarella\n\n- Fresh basil\n\n- Extra virgin olive oil",
		Instructions: "\n\n1. Mix flour and water until incorporated\n\n2. Add yeast and salt, knead for 20 minutes\n\n3. Let rise for 2 hours at room temperature\n\n4. Divide into 250g balls\n\n5. Cold ferment for 24-48 hours\n\n6. Shape by hand, being gentle with the dough\n\n7. Top with crushed tomatoes, torn mozzarella, and basil\n\n8. Cook at 450°C for 60-90 seconds",
		Tips:         "Use a pizza steel or stone preheated for at least 1 hour. The dough should be soft and extensible. For best results, cook in a wood-fired oven.",
	}
)

func init() {
	router.HandleFunc("", renderHomepage)
	router.HandleFunc("recipe/", renderRecipes)

}

func AddRecipe(name, origin, ingredients, instructions, tips string) string {
	if err := validateRecipe(name, ingredients, instructions); err != nil {
		panic(err.Error())
	}

	recipe := &Recipe{
		Name:         name,
		Origin:       origin,
		Author:       std.PrevRealm().Addr(),
		Ingredients:  ingredients,
		Instructions: instructions,
		Tips:         tips,
	}
	recipes = append(recipes, recipe)
	return "Recipe added successfully!"
}

func validateRecipe(name, ingredients, instructions string) error {
	if name == "" {
		return ufmt.Errorf("recipe name cannot be empty")
	}
	if len(ingredients) == 0 {
		return ufmt.Errorf("ingredients cannot be empty")
	}
	if len(instructions) == 0 {
		return ufmt.Errorf("instructions cannot be empty")
	}
	return nil
}

func Follow(addr std.Address) error {
	caller := std.PrevRealm().Addr()
	if !isUser(caller) {
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

func renderRecipes(res *mux.ResponseWriter, req *mux.Request) {
	var b strings.Builder
	b.WriteString("## World Kitchen\n\n------\n\n")
	writeRecipe(&b, margheritaPizza)

	if len(recipes) == 0 {
		b.WriteString("No recipes yet. Be the first to add one!\n")
		res.Write(b.String())
		return
	}

	for _, recipe := range recipes {
		writeRecipe(&b, recipe)
	}
	res.Write(b.String())
}

func writeRecipe(b *strings.Builder, recipe *Recipe) {
	b.WriteString("## " + recipe.Name + "\n")
	b.WriteString("**Author:** " + recipe.Author.String() + "\n\n")
	b.WriteString("**Origin:** " + recipe.Origin + "\n\n")
	b.WriteString("**Ingredients:** " + recipe.Ingredients + "\n\n")

	b.WriteString("**Instructions:** " + recipe.Instructions + "\n\n")

	if recipe.Tips != "" {
		b.WriteString(ufmt.Sprintf("💡 **Tips**: %s\n\n", recipe.Tips))
	}
	b.WriteString("------\n\n")
	b.WriteString("------\n\n")
}

func renderHomepage(res *mux.ResponseWriter, req *mux.Request) {
	var b strings.Builder
	writeNavigation(&b)
	writeProfile(&b)
	res.Write(b.String())
}

func writeProfile(b *strings.Builder) {
	b.WriteString("# Welcome to my Homepage\n\n------\n\n")
	writeGnoArt(b)
	b.WriteString("------\n\n## About Me\n------\n\n")
	b.WriteString(ufmt.Sprintf("![avatar](%s)\n\n", profile.Avatar))
	b.WriteString(profile.AboutMe + "\n\n------\n\n")
	b.WriteString("### Contact\n\n")
	b.WriteString(ufmt.Sprintf("- Email: %s\n", profile.Email))
	b.WriteString(ufmt.Sprintf("- GitHub: [@mous1985](%s)\n", profile.Github))
	b.WriteString(ufmt.Sprintf("- LinkedIn: [Mustapha](%s)\n", profile.LinkedIn))
	b.WriteString("------\n\n")
	b.WriteString(ufmt.Sprintf("👤 Followers: %d", len(profile.Followers)))
}

func writeNavigation(b *strings.Builder) {
	b.WriteString(strings.Join([]string{
		"[Home]()",
		ufmt.Sprintf("[World Kitchen](%s)", Rec),
		"[Hackerspace](https://github.com/gnolang/hackerspace/issues/86#issuecomment-2535795751)",
	}, " | ") + "\n\n------\n\n")
}

func Render(path string) string {
	return router.Render(path)
}
func writeGnoArt(b *strings.Builder) {
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
