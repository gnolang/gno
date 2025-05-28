package place

import (
	"std"
	"time"

	"gno.land/p/demo/svg"
)

func generateSquareSvg(color string) string {
	canvas := svg.Canvas{Width: 30, Height: 30}
	canvas.Append(svg.NewRectangle(0, 0, 30, 30, color))
	return canvas.Base64()
}

func generateSVG() string {
	canvas := svg.Canvas{Width: (int)(cols * 40), Height: (int)(rows * 40)}
	for y := 0; y < int(rows); y++ {
		for x := 0; x < int(cols); x++ {
			canvas.Append(svg.NewRectangle(x*40, y*40, 40, 40, tableColor[y][x].rgb))
		}
	}
	return canvas.Base64()
}

// Place a tile on the tile of choice with the selected color
// Verify that the user is allowed to place a tile
func Submit(x, y uint64, colorName string) {
	user, ok := retrieveMember(std.PreviousRealm().Address())
	if !ok {
		Register()
	}

	if time.Now().Before(user.Timer) {
		panic("wait until " + user.Timer.String())
	}

	foundColor, color := findColor(colorName)

	if !foundColor {
		panic("invalid color")
	}

	tableColor[y][x].rgb = color.rgb
	tableColor[y][x].initial = color.initial

	user.Timer = time.Now().Add(time.Duration(timeRefill) * time.Minute)
	users.Set(std.PreviousRealm().Address().String(), user)
}
