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
		Followers: addrset.Set{},
	}
	router          = mux.NewRouter()
	recipes         []*Recipe
	margheritaPizza = &Recipe{
		Name:         "Authentic Margherita Pizza 🤌",
		Origin:       "Naples, 🇮🇹",
		Author:       config.Address(),
		Ingredients:  "  1kg 00 flour\n 500ml water\n 3g fresh yeast\n 20g sea salt\n San Marzano tomatoes\n Fresh buffalo mozzarella\n Fresh basil\n Extra virgin olive oil",
		Instructions: " Mix flour and water until incorporated\n Add yeast and salt, knead for 20 minutes\n Let rise for 2 hours at room temperature\n Divide into 250g balls\n Cold ferment for 24-48 hours\n Shape by hand, being gentle with the dough\n Top with crushed tomatoes, torn mozzarella, and basil\n Cook at 450°C for 60-90 seconds",
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
// TODO:any user can follow and to be followed by any other user
// TODO: add a function to unfollow

func Follow(addr std.Address) error {
	caller := std.PrevRealm().Addr()
	if !isUser(caller) {
		return config.ErrorUnauthorized
	}
	if profile.Followers.Has(addr) {
		return ufmt.Errorf("address %s is already following", addr)
	}
	profile.Followers.Add(addr) //can't add the same address twice
	return nil
}

func isUser(addr std.Address) bool {
	return !isAuthorized(addr)
}
func isAuthorized(addr std.Address) bool {
	return addr == config.Address() || addr == config.Backup()
}

func renderRecipes(res *mux.ResponseWriter, req *mux.Request) {
	var b strings.Builder
	b.WriteString("## World Kitchen\n\n------\n\n")

	b.WriteString(margheritaPizza.Render())

	if len(recipes) == 0 {
		b.WriteString("No recipes yet. Be the first to add one!\n")
		res.Write(b.String())
		return
	}

	for _, recipe := range recipes {
		b.WriteString(recipe.Render())
	}

	res.Write(b.String())
}

func (r Recipe) Render() string {
	var b strings.Builder
	b.WriteString(md.H2(r.Name))
	b.WriteString(md.Bold("Author:") + "\n" + r.Author.String() + "\n\n")
	b.WriteString(md.Bold("Origin:") + "\n" + r.Origin + "\n\n")
	b.WriteString(md.Bold("Ingredients:") + "\n" + md.BulletList(strings.Split(r.Ingredients, "\n")) + "\n\n")
	b.WriteString(md.Bold("Instructions:") + "\n" + md.OrderedList(strings.Split(r.Instructions, "\n")) + "\n\n")

	if r.Tips != "" {
		b.WriteString(md.Italic("💡 Tips:"+"\n"+r.Tips) + "\n\n")
	}

	b.WriteString(md.HorizontalRule() + "\n")
	return b.String()
}

func renderHomepage(res *mux.ResponseWriter, req *mux.Request) {
	var b strings.Builder
	writeNavigation(&b)
	b.WriteString(profile.Render())
	res.Write(b.String())
}

func (p Profile) Render() string {
	var b strings.Builder

	b.WriteString(md.H1("Welcome to my Homepage") + "\n\n" + md.HorizontalRule() + "\n\n")
	writeGnoArt(&b)
	b.WriteString(md.HorizontalRule() + "\n\n" + md.H2("About Me") + "\n\n")
	b.WriteString(md.Image("avatar", p.Avatar) + "\n\n")
	b.WriteString(p.AboutMe + "\n\n" + md.HorizontalRule() + "\n\n")
	b.WriteString(md.H3("Contact") + "\n\n")
	b.WriteString(md.BulletList([]string{
		"Email: " + p.Email,
		"GitHub: " + md.Link("@mous1985", p.Github),
		"LinkedIn: " + md.Link("Mustapha", p.LinkedIn),
	}))
	b.WriteString("\n\n" + md.Bold("👤 Followers: ") + strconv.Itoa(p.Followers.Size()))

	return b.String()
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
		"      -==++.                                                                  ",
		"     *@@@@=                                     @-                          -@",
		"    #@@@@@:       -==-.-- :-::===:   .-++-.     @-   .===:.- .-.-==-   .===:=@",
		"   #@@@@@@@:    -@@%**%@@ #@@#*#@@- *@@**@@*    @-  +%=::-*@ +@=-:-@* +%=::-*@",
		"  +@%#**#%@@    %@+   :@@ *@+   #@=+@%    %@+   @= :@:    -@ +%    +%.@:    -@",
		"  -:       -    *@%:..+@@ *@+   #@=-@@:  :@@=   @- .@=    =@ +@    *%.@=    =@",
		"  --:==+=-:=.    =%@%#*@@ *@+   #@+ =%@%%@%= #* %#=.:%*===*@ +%    +% -%*===*@",
		"  -++++=++++.    =-:::*@#  .     .    .::.   ..  ::   .::  .  .         .::  .",
		"   .-=+++=:     .*###%#=                                                      ",
		"      ::                                                                      ",
	} {
		b.WriteString(line + "\n")
	}
	b.WriteString("```\n------")
}
