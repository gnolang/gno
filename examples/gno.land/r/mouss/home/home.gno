package home

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/addrset"
	"gno.land/p/moul/md"
	"gno.land/r/mouss/config"
)

// Profile represents my personal profile information.
type Profile struct {
	AboutMe   string
	Avatar    string
	Email     string
	Github    string
	LinkedIn  string
	Followers addrset.Set // Set of followers addresses.
}

// Recipe represents a cooking recipe with its details.
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

// init initializes the router with the homepage and recipe routes.
func init() {
	router.HandleFunc("", renderHomepage)
	router.HandleFunc("recipe/", renderRecipes)
}

// AddRecipe adds a new recipe in recipe page by users
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

// validateRecipe checks if the provided recipe details are valid.
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

// Follow allows a users to follow my home page.
// It checks if the caller is a valid user and if the address is already being followed.
// If the caller is not authorized, it returns an error.
// If the address is already being followed, it returns an error.
// Otherwise, it adds the address to the list of followers and returns nil.
//TODO:any user can follow and to be followed by any other user
//TODO: add a function to unfollow

func Follow(addr std.Address) error {
	caller := std.PrevRealm().Addr()
	if !isUser(caller) {
		return config.ErrorUnauthorized
	}
	if profile.Followers.Has(addr) {
		return ufmt.Errorf("address %s is already following", addr)
	}
	profile.Followers.Add(addr)
	return nil
}

func isUser(addr std.Address) bool {
	return !isAuthorized(addr) && !profile.Followers.Has(addr)
}
func isAuthorized(addr std.Address) bool {
	return addr == config.Address() || addr == config.Backup()
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
	b.WriteString(md.H2(recipe.Name))
	b.WriteString(md.Bold("Author: ") + recipe.Author.String() + "\n\n")
	b.WriteString(md.Bold("Origin: ") + recipe.Origin + "\n\n")
	b.WriteString(md.Bold("Ingredients:") + "\n" + md.BulletList(strings.Split(recipe.Ingredients, "\n")) + "\n")
	b.WriteString(md.Bold("Instructions:") + "\n" + md.OrderedList(strings.Split(recipe.Instructions, "\n")) + "\n")

	if recipe.Tips != "" {
		b.WriteString(md.Italic("💡 Tips: "+recipe.Tips) + "\n\n")
	}

	b.WriteString(md.HorizontalRule() + "\n")
}

func renderHomepage(res *mux.ResponseWriter, req *mux.Request) {
	var b strings.Builder
	writeNavigation(&b)
	writeProfile(&b)
	res.Write(b.String())
}

func writeProfile(b *strings.Builder) {
	b.WriteString(md.H1("Welcome to my Homepage") + "\n\n" + md.HorizontalRule() + "\n\n")
	writeGnoArt(b)
	b.WriteString(md.HorizontalRule() + "\n\n" + md.H2("About Me") + "\n\n")
	b.WriteString(md.Image("avatar", profile.Avatar) + "\n\n")
	b.WriteString(profile.AboutMe + "\n\n" + md.HorizontalRule() + "\n\n")
	b.WriteString(md.H3("Contact") + "\n\n")
	b.WriteString(md.BulletList([]string{
		"Email: " + profile.Email,
		"GitHub: " + md.Link("@mous1985", profile.Github),
		"LinkedIn: " + md.Link("Mustapha", profile.LinkedIn),
	}))
	b.WriteString("\n\n" + md.Bold("👤 Followers: ") + strconv.Itoa(profile.Followers.Size()))
}
func writeNavigation(b *strings.Builder) {
	navItems := []string{
		md.Link("Home", ""),
		md.Link("World Kitchen", Rec),
		md.Link("Hackerspace", "https://github.com/gnolang/hackerspace/issues/86#issuecomment-2535795751"),
	}
	b.WriteString(strings.Join(navItems, " | ") + "\n\n" + md.HorizontalRule() + "\n\n")
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
