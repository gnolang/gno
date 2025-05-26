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

// deepCopy creates a deep copy of a Pokémon, including its attacks.
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

// getDefault retrieves the default Pokémon data from pokemonList by it's name
func (p *Pokemon) getDefault() *Pokemon {
	for _, d := range pokemonList {
		if p.name == d.name {
			return &d
		}
	}
	return nil
}

// getId returns the index of the Pokémon in pokemonList by name, or -1 if not found
func (p *Pokemon) getId() int {
	for i, d := range pokemonList {
		if p.name == d.name {
			return i
		}
	}
	return -1
}

// pokemonList contains predefined Pokémon with their attributes and attacks
var (
	pokemonList = []Pokemon{
		{
			"Pikachu",
			"https://raw.githubusercontent.com/HybridShivam/Pokemon/refs/heads/master/assets/images/025.png",
			60,
			[]Attack{
				{"Thunderbolt", 50, 2},
				{"Tackle", 10, 10},
				{"Iron Tail", 30, 5},
			},
		},
		{
			"Bulbasaur",
			"https://raw.githubusercontent.com/HybridShivam/Pokemon/refs/heads/master/assets/images/001.png",
			80,
			[]Attack{
				{"Razor Leaf", 25, 6},
				{"False Swipe", 15, 10},
				{"Vine Whip", 45, 2},
			},
		},
		{
			"Charmander",
			"https://raw.githubusercontent.com/HybridShivam/Pokemon/refs/heads/master/assets/images/004.png",
			70,
			[]Attack{
				{"Flamethrower", 40, 3},
				{"Scratch", 10, 15},
				{"Fire Spin", 35, 4},
			},
		},
		{
			"Squirtle",
			"https://raw.githubusercontent.com/HybridShivam/Pokemon/refs/heads/master/assets/images/007.png",
			75,
			[]Attack{
				{"Water Gun", 30, 6},
				{"Tackle", 10, 12},
				{"Bubble Beam", 40, 3},
			},
		},
		{
			"Eevee",
			"https://raw.githubusercontent.com/HybridShivam/Pokemon/refs/heads/master/assets/images/133.png",
			65,
			[]Attack{
				{"Quick Attack", 20, 10},
				{"Swift", 25, 8},
				{"Bite", 30, 5},
			},
		},
		{
			"Jigglypuff",
			"https://raw.githubusercontent.com/HybridShivam/Pokemon/refs/heads/master/assets/images/039.png",
			85,
			[]Attack{
				{"Sing", 0, 5},
				{"Pound", 15, 12},
				{"Double Slap", 25, 6},
			},
		},
		{
			"Meowth",
			"https://raw.githubusercontent.com/HybridShivam/Pokemon/refs/heads/master/assets/images/052.png",
			60,
			[]Attack{
				{"Scratch", 15, 12},
				{"Bite", 25, 7},
				{"Pay Day", 30, 4},
			},
		},
		{
			"Psyduck",
			"https://raw.githubusercontent.com/HybridShivam/Pokemon/refs/heads/master/assets/images/054.png",
			70,
			[]Attack{
				{"Confusion", 35, 5},
				{"Water Gun", 30, 6},
				{"Zen Headbutt", 40, 3},
			},
		},
		{
			"Snorlax",
			"https://raw.githubusercontent.com/HybridShivam/Pokemon/refs/heads/master/assets/images/143.png",
			160,
			[]Attack{
				{"Body Slam", 40, 5},
				{"Lick", 20, 10},
				{"Hyper Beam", 60, 2},
			},
		},
		{
			"Gengar",
			"https://raw.githubusercontent.com/HybridShivam/Pokemon/refs/heads/master/assets/images/094.png",
			80,
			[]Attack{
				{"Shadow Ball", 45, 3},
				{"Lick", 20, 10},
				{"Dark Pulse", 35, 4},
			},
		},
	}
)
