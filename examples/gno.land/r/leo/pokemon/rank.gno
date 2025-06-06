package pokemon

import (
	"sort"

	"gno.land/p/demo/mux"
	"gno.land/p/moul/md"
)

type playerEntry struct {
	adress     string
	pokedollar int
}

// playerList is a slice of playerEntry for sorting players by Pokedollars
type playerList []playerEntry

func (p playerList) Len() int {
	return len(p)
}

// Less compares two players by Pokedollar amount (descending order)
func (p playerList) Less(i, j int) bool {
	return p[i].pokedollar > p[j].pokedollar
}

// Swap exchanges two players in the playerList for sorting
func (p playerList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// renderPokemonRank renders the ranking page with players sorted by Pokedollars
func renderPokemonRank(rend *mux.ResponseWriter, req *mux.Request) {
	rend.Write(md.H2("Ranking"))

	var players playerList
	pokemonDAO.Iterate("", "", func(addr string, user any) bool {
		players = append(players, playerEntry{
			adress:     addr,
			pokedollar: int(user.(*UserData).pokedollar),
		})
		return false
	})

	sort.Sort(players)

	var playerOrdered []string
	for _, u := range players {
		if u.adress == req.Query.Get("user") {
			playerOrdered = append(playerOrdered, "-> You: "+u.adress+" : "+renderPokedollar(u.pokedollar))
		} else {
			playerOrdered = append(playerOrdered, u.adress+" : "+renderPokedollar(u.pokedollar))
		}
	}
	rend.Write(md.OrderedList(playerOrdered) + nl)
	rend.Write(md.Link("Go back to lobby", atmPath+":lobby?user="+req.Query.Get("user")))
	return
}
