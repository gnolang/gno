package gnominers

// "strconv"
// "gno.land/p/pol/notrandom"

type GemType uint

const (
	GemRock GemType = iota
	GemQuartz
	GemAmethyst
	GemTopaz
	GemOpal
	GemSapphire
	GemEmerald
	GemRuby
	GemDiamond
	GemAlexandrite
	_GemEnd
)

var gemLootTable = map[GemType]float64{
	GemRock:        1,
	GemQuartz:      0.2,
	GemAmethyst:    0.1,
	GemTopaz:       0.05,
	GemOpal:        0.021,
	GemSapphire:    0.01,
	GemEmerald:     0.00723,
	GemRuby:        0.0069,
	GemDiamond:     0.005,
	GemAlexandrite: 0.001,
}

var gemPriceCaratTable = map[GemType]float64{
	GemRock:        0.01,
	GemQuartz:      1,
	GemAmethyst:    5,
	GemTopaz:       25,
	GemOpal:        800,
	GemSapphire:    2300,
	GemEmerald:     6500,
	GemRuby:        1500,
	GemDiamond:     15000,
	GemAlexandrite: 12000,
}

var GemNames = []string{
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

// func randomGem(m *notrandom.Machine) (GemType, float64) {
func randomGem() (GemType, float64) {
	total := 0.
	for _, v := range gemLootTable {
		total += v
	}
	// r := MachineRandFloat()
	r := RandomMachine.RandFloat()
	value := r * total
	for k, v := range gemLootTable {
		if value < v {
			return k, r
		} else {
			value -= v
		}
	}
	return GemRock, r
}
