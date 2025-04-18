package pokemon

import (
	"net/url"
	"std"
	"strings"

	"gno.land/p/demo/mux"
	"gno.land/p/leo/randomizer"
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
	rand        = randomizer.NewRandomizer()
	atmPath     = std.CurrentRealm().PkgPath()[len(std.ChainDomain()):]
	pokemonPath = map[RenderPage]mux.HandlerFunc{
		pokemonLogin:        renderPokemonLogin,
		pokemonLobby:        renderPokemonLobby,
		pokemonTeamSelector: renderPokemonTeam,
		pokemonFight:        renderPokemonFight,
		pokemonShop:         renderPokemonShop,
		pokemonRank:         renderPokemonRank,
	}
)

func init() {
	hor.Register("Pokémon", "Come play some Pokémon's fights")
}

func Render(path string) (rend string) {
	router := mux.NewRouter()
	rend += md.H1("Pokémon")
	_, rawQuery, _ := strings.Cut(path, "?")
	query, _ := url.ParseQuery(rawQuery)

	if pokemonDAO.IsMember(std.Address(query.Get("user"))) {
		for p, f := range pokemonPath {
			router.HandleFunc(string(p), f)
		}
	}
	router.NotFoundHandler = renderPokemonLogin
	rend += router.Render(path)
	return
}
