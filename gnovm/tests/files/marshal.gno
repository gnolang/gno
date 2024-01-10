package main

import (
	"fmt"

	"gno.land/p/demo/json"
)

func main() {
	character := map[string]interface{}{
		"name": "Eldon",
		"class": "Wizard",
		"race": "Elf",
		"level": 5,
		"stats": map[string]interface{}{
			"strength": 8.5,
			"dexterity": 14,
			"constitution": 10,
			"intelligence": 18,
			"wisdom": 12,
			"charisma": 11,
		},
		"hitPoints": 27,
		"alignment": "Neutral Good",
		"skills": map[string]interface{}{
			"arcana": 7,
			"history": 7,
			"investigation": 6,
			"nature": 5,
			"religion": 5,
		},
		"spells": []interface{}{"Magic Missile", "Fireball", "Invisibility", "Shield"},
		"inventory": map[string]interface{}{
			"weapons": []interface{}{"Quarterstaff", "Dagger"},
			"armor": "Mage Armor",
			"items": []interface{}{"Potion of Healing", "Spellbook"},
			"gold": 120,
		},
		"background": map[string]interface{}{
			"story": "Raised by a secret society of wizards...",
			"traits": []interface{}{"Curious", "Cautious"},
			"ideals": "Knowledge is the path to power and domination.",
			"bonds": "Must protect ancient wizardry tome.",
			"flaws": "Can't resist a new mystery.",
		},
	}

	json, err := json.Marshal(character)
	if err != nil { panic(err) }
	fmt.Println(json)
}
