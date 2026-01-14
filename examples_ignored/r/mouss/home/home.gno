package home

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/addrset"
	"gno.land/p/moul/md"
	"gno.land/r/leon/hor"
	"gno.land/r/mouss/config"
)

// Profile represents my personal profile information.
type Profile struct {
	AboutMe   string
	Avatar    string
	Email     string
	Github    string
	LinkedIn  string
	Followers *addrset.Set // Set of followers addresses.
}

// Recipe represents a cooking recipe with its details.
type Recipe struct {
	Name         string
	Origin       string
	Author       address
	Ingredients  string
	Instructions string
	Tips         string
}

const (
	realmURL = "/r/mouss/home"
	rec      = realmURL + ":recipe/"
	gnoArt   = `
        -==++.
	     *@@@@=                                     @-                          -@
	    #@@@@@:       -==-.-- :-::===:   .-++-.     @-   .===:.- .-.-==-   .===:=@
       #@@@@@@@:    -@@%**%@@ #@@#*#@@- *@@**@@*    @-  +%=::-*@ +@=-:-@* +%=::-*@
      +@%#**#%@@    %@+   :@@ *@+   #@=+@%    %@+   @= :@:    -@ +%    +%.@:    -@
      -:       -    *@%:..+@@ *@+   #@=-@@:  :@@=   @- .@=    =@ +@    *%.@=    =@
      --:==+=-:=.    =%@%#*@@ *@+   #@+ =%@%%@%= #* %#=.:%*===*@ +%    +% -%*===*@
      -++++=++++.    =-:::*@#  .     .    .::.   ..  ::   .::  .  .         .::  .
       .-=+++=:     .*###%#=
	      ::
`
)

var (
	router          = mux.NewRouter()
	profile         Profile
	recipes         = avl.NewTree()
	margheritaPizza *Recipe
)

// init initializes the router with the home page and recipe routes
// sets up my profile information, and my recipe
// and registers the home page in the hall of fame.
func init() {
	router.HandleFunc("", renderHomepage)
	router.HandleFunc("recipe/", renderRecipes)
	router.HandleFunc("recipe/{name}", renderRecipe)
	profile = Profile{
		AboutMe:   "👋 I'm Mustapha, a contributor to gno.land project from France. I'm passionate about coding, exploring new technologies, and contributing to open-source projects. Besides my tech journey, I'm also a pizzaiolo 🍕 who loves cooking and savoring good food.",
		Avatar:    "https://github.com/mous1985/assets/blob/master/avatar.png?raw=true",
		Email:     "mustapha.benazzouz@outlook.fr",
		Github:    "https://github.com/mous1985",
		LinkedIn:  "https://www.linkedin.com/in/mustapha-benazzouz-88646887/",
		Followers: &addrset.Set{},
	}
	margheritaPizza = &Recipe{
		Name:         "Authentic Margherita Pizza 🤌",
		Origin:       "Naples, 🇮🇹",
		Author:       config.OwnableMain.Owner(),
		Ingredients:  "  1kg 00 flour\n 500ml water\n 3g fresh yeast\n 20g sea salt\n San Marzano tomatoes\n Fresh buffalo mozzarella\n Fresh basil\n Extra virgin olive oil",
		Instructions: " Mix flour and water until incorporated\n Add yeast and salt, knead for 20 minutes\n Let rise for 2 hours at room temperature\n Divide into 250g balls\n Cold ferment for 24-48 hours\n Shape by hand, being gentle with the dough\n Top with crushed tomatoes, torn mozzarella, and basil\n Cook at 450°C for 60-90 seconds",
		Tips:         "Use a pizza steel or stone preheated for at least 1 hour. The dough should be soft and extensible. For best results, cook in a wood-fired oven.",
	}
	hor.Register("Mouss's Home Realm", "")
}

// AddRecipe adds a new recipe in recipe page by users
func AddRecipe(name, origin, ingredients, instructions, tips string) string {
	if err := validateRecipe(name, ingredients, instructions); err != nil {
		panic(err)
	}
	recipe := &Recipe{
		Name:         name,
		Origin:       origin,
		Author:       std.PreviousRealm().Address(),
		Ingredients:  ingredients,
		Instructions: instructions,
		Tips:         tips,
	}
	recipes.Set(name, recipe)
	return "Recipe added successfully"
}

func UpdateAboutMe(about string) error {
	if !config.IsAuthorized(std.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}
	profile.AboutMe = about
	return nil
}

func UpdateAvatar(avatar string) error {
	if !config.IsAuthorized(std.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}
	profile.Avatar = avatar
	return nil
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
// If the caller is admin it returns error.
func Follow() error {
	caller := std.PreviousRealm().Address()

	if caller == config.OwnableMain.Owner() {
		return ufmt.Errorf("you cannot follow yourself")
	}
	if profile.Followers.Add(caller) {
		return nil
	}
	return ufmt.Errorf("you are already following")
}

// Unfollow allows a user to unfollow my home page.
func Unfollow() error {
	caller := std.PreviousRealm().Address()

	if profile.Followers.Remove(caller) {
		return nil
	}
	return ufmt.Errorf("you are not following")
}

// renderRecipes renders the list of recipes.
func renderRecipes(res *mux.ResponseWriter, req *mux.Request) {
	var out string
	out += Header()
	out += "## World Kitchen\n\n------\n\n"

	// Link to margarita pizza recipe
	out += "### Available Recipes:\n\n"
	out += "* " + md.Link(margheritaPizza.Name, rec+"margheritaPizza") + "By : " + string(margheritaPizza.Author) + "\n"

	// The list of all other recipes with clickable links
	if recipes.Size() > 0 {
		recipes.Iterate("", "", func(key string, value interface{}) bool {
			recipe := value.(*Recipe)
			out += "* " + md.Link(recipe.Name, rec+recipe.Name) + " By : " + recipe.Author.String() + "\n"
			return false // continue iterating
		})
		out += "\n------\n\n"
	} else {
		out += "\nNo additional recipes yet. Be the first to add one!\n"
	}
	res.Write(out)
}

// renderRecipe renders the recipe details.
func renderRecipe(res *mux.ResponseWriter, req *mux.Request) {
	name := req.GetVar("name")
	if name == "margheritaPizza" {
		res.Write(margheritaPizza.Render())
		return
	}
	value, exists := recipes.Get(name)
	if !exists {
		res.Write("Recipe not found")
		return
	}
	recipe := value.(*Recipe)
	res.Write(recipe.Render())
}

func (r Recipe) Render() string {
	var out string
	out += Header()
	out += md.H2(r.Name)
	out += md.Bold("Author:") + "\n" + r.Author.String() + "\n\n"
	out += md.Bold("Origin:") + "\n" + r.Origin + "\n\n"
	out += md.Bold("Ingredients:") + "\n" + md.BulletList(strings.Split(r.Ingredients, "\n")) + "\n\n"
	out += md.Bold("Instructions:") + "\n" + md.OrderedList(strings.Split(r.Instructions, "\n")) + "\n\n"
	if r.Tips != "" {
		out += md.Italic("💡 Tips:"+"\n"+r.Tips) + "\n\n"
	}
	out += md.HorizontalRule() + "\n"
	return out
}

func renderHomepage(res *mux.ResponseWriter, req *mux.Request) {
	var out string
	out += Header()
	out += profile.Render()
	res.Write(out)
}

func (p Profile) Render() string {
	var out string
	out += md.H1("Welcome to my Homepage") + "\n\n" + md.HorizontalRule() + "\n\n"
	out += "```\n"
	out += gnoArt
	out += "```\n------"
	out += md.HorizontalRule() + "\n\n" + md.H2("About Me") + "\n\n"
	out += md.Image("avatar", p.Avatar) + "\n\n"
	out += p.AboutMe + "\n\n" + md.HorizontalRule() + "\n\n"
	out += md.H3("Contact") + "\n\n"
	out += md.BulletList([]string{
		"Email: " + p.Email,
		"GitHub: " + md.Link("@mous1985", p.Github),
		"LinkedIn: " + md.Link("Mustapha", p.LinkedIn),
	})
	out += "\n\n" + md.Bold("👤 Followers: ") + strconv.Itoa(p.Followers.Size())
	return out
}

func Header() string {
	navItems := []string{
		md.Link("Home", realmURL),
		md.Link("World Kitchen", rec),
		md.Link("Hackerspace", "https://github.com/gnolang/hackerspace/issues/86#issuecomment-2535795751"),
	}
	return strings.Join(navItems, " | ") + "\n\n" + md.HorizontalRule() + "\n\n"
}

func Render(path string) string {
	return router.Render(path)
}
