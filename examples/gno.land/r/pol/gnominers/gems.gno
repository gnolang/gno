package gnominers

// "strconv"
// "gno.land/p/pol/notrandom"

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
	gemRock:        1,
	gemQuartz:      0.2,
	gemAmethyst:    0.1,
	gemTopaz:       0.05,
	gemOpal:        0.021,
	gemSapphire:    0.01,
	gemEmerald:     0.00723,
	gemRuby:        0.0069,
	gemDiamond:     0.005,
	gemAlexandrite: 0.001,
}

var gemPriceCaratTable = map[gemType]float64{
	gemRock:        0.01,
	gemQuartz:      1,
	gemAmethyst:    5,
	gemTopaz:       25,
	gemOpal:        800,
	gemSapphire:    2300,
	gemEmerald:     6500,
	gemRuby:        1500,
	gemDiamond:     15000,
	gemAlexandrite: 12000,
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

// func randomGem(m *notrandom.Machine) (gemType, float64) {
func randomGem() (gemType, float64) {
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
