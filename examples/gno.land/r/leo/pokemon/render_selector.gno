package pokemon

import (
	"math/rand"
	"net/url"
	"std"
	"strings"
	"time"

	"gno.land/p/demo/mux"
	"gno.land/p/moul/md"
	"gno.land/r/leon/hor"
)

type RenderPage string

const (
	pokemonLobby        RenderPage = "lobby"
	pokemonLogin        RenderPage = "login"
	pokemonTeamSelector RenderPage = "team"
	pokemonFight        RenderPage = "fight"
	pokemonShop         RenderPage = "shop"
	pokemonRank         RenderPage = "rank"
	nl                             = "\n\n"
)

var (
	// time-based seed for combat enemy IA (random issou)
	r           = rand.New(rand.NewPCG(uint64(time.Now().Unix()), uint64(time.Now().Unix())))
	atmPath     = std.CurrentRealm().PkgPath()[len(std.ChainDomain()):]
	pokemonPath = map[RenderPage]mux.HandlerFunc{
		pokemonLogin:        renderPokemonLogin, // login page for user selection
		pokemonLobby:        renderPokemonLobby, // lobby page for game navigation
		pokemonTeamSelector: renderPokemonTeam,  // team selection page for choosing Pokémon
		pokemonFight:        renderPokemonFight, // fight page for Pokémon battles
		pokemonShop:         renderPokemonShop,  // shop page for purchasing items
		pokemonRank:         renderPokemonRank,  // ranking page for player standings
	}
)

func init() {
	cross(hor.Register)("Pokémon", "Come play some Pokémon's fights")
}

// render the Pokémon game pages based on the provided path and user query
func Render(path string) (rend string) {
	router := mux.NewRouter()
	rend += md.H1("Pokémon")
	_, rawQuery, _ := strings.Cut(path, "?")
	query, _ := url.ParseQuery(rawQuery)

	if pokemonDAO.Has((query.Get("user"))) {
		for p, f := range pokemonPath {
			router.HandleFunc(string(p), f)
		}
	}
	router.NotFoundHandler = renderPokemonLogin
	rend += router.Render(path)
	return
}
