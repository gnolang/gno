package main

import (
	"fmt"

	"gno.land/p/demo/json"
)

// demo for value retrieval from JSON data
func main() {
	data := []byte(`
	{
		"character": {
			"race": "Tenebrine",
			"species": "Raccoon",
			"class": "Thief",
			"name": "Whisker"
			"stats": {
				"strength": 8,
				"dexterity": 18,
				"constitution": 12,
				"intelligence": 11,
				"wisdom": 12,
				"charisma": 14
			},
		},
		"skills": {
			"acrobatics": 4,
			"animal_handling": 1,
			"arcana": 1,
			"athletics": 2,
			"deception": 5,
			"history": 1,
			"insight": 3,
			"intimidation": 2,
			"investigation": 4,
			"medicine": 1,
			"nature": 1,
			"perception": 4,
			"performance": 3,
			"persuasion": 3,
			"religion": 1,
			"sleight_of_hand": 6,
			"stealth": 5,
			"survival": 3
		},
		"equipment": {
			"weapons": ["Daggers", "Short sword"],
			"armor": "Leather armor",
			"items": ["Thieves' tools", "Backpack", "Rations (5 days)", "Lock picks", "Disguise kit"],
			"gold": 200
		},
		"backstory": "Raised on the streets of the shadowy Tenebrine city, ${player} learned to survive by his wits and agility.\nSkilled in the art of stealth and deception, he mastered the craft of thievery to claim what the streets denied him.\nKnown only by his alias 'Shadowpaw', he now seeks greater challenges and wealth beyond the alleyways of his upbringing."
	}
	`)

	// `json.Get` function takes a JSON and a path of keys (must enter each key name separately) to retrieve the value.
	if val, typ, _, err := json.Get(data, "character", "name"); err != nil {
		panic(err)
	} else {
		fmt.Printf("name: %s type: %v\n", val, typ.String())
	}

	if ab, typ, _, err := json.Get(data, "character", "stats"); err != nil {
		panic(err)
	} else {
		fmt.Printf("stats: %s, type: %v\n\n", ab, typ.String())
	}

	if stealth, typ, _, err := json.Get(data, "skills", "stealth"); err != nil {
		panic(err)
	} else {
		fmt.Printf("stealth: %s, type: %v\n\n", stealth, typ.String())
	}

	if weapons, typ, _, err := json.Get(data, "equipment", "weapons"); err != nil {
		panic(err)
	} else {
		fmt.Printf("weapons: %s, type: %v\n\n", weapons, typ.String())
	}

	if armor, typ, _, err := json.Get(data, "equipment", "armor"); err != nil {
		panic(err)
	} else {
		fmt.Printf("armor: %s, type: %v\n\n", armor, typ.String())
	}

	if backstory, typ, _, err := json.Get(data, "backstory"); err != nil {
		panic(err)
	} else {
		fmt.Printf("backstory: %s, type: %v\n\n", backstory, typ.String())
	}
}