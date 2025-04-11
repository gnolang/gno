package pokemon

type Attack struct {
	name   string
	damage int
	pc     int
}

type Pokemon struct {
	name      string
	assetPath string
	pv        int
	attacks   []Attack
}

type ItemType int

type Item struct {
	price       int
	name        string
	description string
	applyer     func(*Pokemon)
}

const (
	teamSize             = 6
	SmallPotion ItemType = iota
	BigPotion
	MaxPotion
	Revive
	MaxRevive
	Elixer
	MaxElixer
)

func (p *Pokemon) deepCopy() Pokemon {
	copiedAttacks := make([]Attack, len(p.attacks))
	copy(copiedAttacks, p.attacks)
	return Pokemon{
		name:      p.name,
		assetPath: p.assetPath,
		pv:        p.pv,
		attacks:   copiedAttacks,
	}
}

func (p *Pokemon) getDefault() *Pokemon {
	for _, d := range pokemonList {
		if p.name == d.name {
			return &d
		}
	}
	return nil
}

func (p *Pokemon) getId() int {
	for i, d := range pokemonList {
		if p.name == d.name {
			return i
		}
	}
	return -1
}

var (
	pokemonList = []Pokemon{
		{
			"Pikachu",
			"https://www.pokemon.com/static-assets/content-assets/cms2/img/pokedex/full/025.png",
			60,
			[]Attack{
				{"Thunderbolt", 50, 2},
				{"Tackle", 10, 10},
				{"Iron Tail", 30, 5},
			},
		},
		{
			"Bulbasaur",
			"https://www.pokemon.com/static-assets/content-assets/cms2/img/pokedex/full/001.png",
			80,
			[]Attack{
				{"Razor Leaf", 25, 6},
				{"False Swipe", 15, 10},
				{"Vine Whip", 45, 2},
			},
		},
		{
			"Charmander",
			"https://www.pokemon.com/static-assets/content-assets/cms2/img/pokedex/full/004.png",
			70,
			[]Attack{
				{"Flamethrower", 40, 3},
				{"Scratch", 10, 15},
				{"Fire Spin", 35, 4},
			},
		},
		{
			"Squirtle",
			"https://www.pokemon.com/static-assets/content-assets/cms2/img/pokedex/full/007.png",
			75,
			[]Attack{
				{"Water Gun", 30, 6},
				{"Tackle", 10, 12},
				{"Bubble Beam", 40, 3},
			},
		},
		{
			"Eevee",
			"https://www.pokemon.com/static-assets/content-assets/cms2/img/pokedex/full/133.png",
			65,
			[]Attack{
				{"Quick Attack", 20, 10},
				{"Swift", 25, 8},
				{"Bite", 30, 5},
			},
		},
		{
			"Jigglypuff",
			"https://www.pokemon.com/static-assets/content-assets/cms2/img/pokedex/full/039.png",
			85,
			[]Attack{
				{"Sing", 0, 5},
				{"Pound", 15, 12},
				{"Double Slap", 25, 6},
			},
		},
	}
)
