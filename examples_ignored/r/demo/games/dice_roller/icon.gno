package dice_roller

import (
	"strconv"
)

// diceIcon returns an icon of the dice roll
func diceIcon(roll int) string {
	switch roll {
	case 1:
		return "🎲1"
	case 2:
		return "🎲2"
	case 3:
		return "🎲3"
	case 4:
		return "🎲4"
	case 5:
		return "🎲5"
	case 6:
		return "🎲6"
	default:
		return "❓"
	}
}

// resultIcon returns the icon representing the result of a game
func resultIcon(result int) string {
	switch result {
	case ongoing:
		return "🔄"
	case win:
		return "🏆"
	case loss:
		return "❌"
	case draw:
		return "🤝"
	default:
		return "❓"
	}
}

// rankIcon returns the icon for a player's rank
func rankIcon(rank int) string {
	switch rank {
	case 1:
		return "🥇"
	case 2:
		return "🥈"
	case 3:
		return "🥉"
	default:
		return strconv.Itoa(rank)
	}
}
