package gnominers

import (
	"strconv"
	// "gno.land/p/pol/notrandom"
)

type gemType uint
const (
	gemRock gemType = iota
	gemQuartz
	gemAmethyst
	gemTopaz
	gemOpal
	gemSapphire
	gemEmerald
	gemRuby
	gemDiamond
	gemAlexandrite
	_gemEnd
)

var gemLootTable = map[gemType]float64{
	gemRock: 			1,
	gemQuartz:			0.2,
	gemAmethyst: 		0.1,
	gemTopaz: 			0.05,
	gemOpal: 			0.042,
	gemSapphire:		0.01,
	gemEmerald: 		0.00723,
	gemRuby: 			0.0069,
	gemDiamond: 		0.005,
	gemAlexandrite: 	0.001,
}

var gemPriceCaratTable = map[gemType]float64{
	gemRock: 			0.01,
	gemQuartz: 			1,
	gemAmethyst: 		5,
	gemTopaz: 			25,
	gemOpal: 			2300,
	gemSapphire: 		800,
	gemEmerald: 		6500,
	gemRuby: 			1500,
	gemDiamond: 		15000,
	gemAlexandrite: 	12000,
}

var gemNames = []string{
	"rock",
	"quartz",
	"amethyst",
	"topaz",
	"opal",
	"sapphire",
	"emerald",
	"ruby",
	"diamond",
	"alexandrite",
}

func debugGem() string {
	total := 0.
	for _, v := range gemLootTable {
		total += v
	}
	content := strconv.FormatFloat(total, 'f', -1, 64)
	content += ", rarest is " + strconv.FormatFloat(gemLootTable[gemDiamond] / total, 'f', -1, 64)
	content += ", lest rare is " + strconv.FormatFloat(gemLootTable[gemRock] / total, 'f', -1, 64)
	return content
}

// func RandomGem(m *notrandom.Machine) (gemType, float64) {
func RandomGem() (gemType, float64) {
	total := 0.
	for _, v := range gemLootTable {
		total += v
	}
	// r := MachineRandFloat()
	r := randomMachine.RandFloat()
	value := r * total
	for k, v := range gemLootTable {
		if value < v {
			return k, r
		} else {
			value -= v
		}
	}
	return gemRock, r
}
