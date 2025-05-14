package place

import (
	"encoding/base64"
	"std"
	"time"

	"gno.land/p/demo/svg"
	"gno.land/p/demo/ufmt"
)

func generateSquareSvg(color string) string {
	canvas := svg.Canvas{Width: 30, Height: 30}
	canvas.DrawRectangle(0, 0, 30, 30, color)
	return base64.StdEncoding.EncodeToString([]byte(canvas.String()))
}

func GenerateSVG() string {
	out := ufmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, cols*40, rows*40)
	for y := 0; y < int(rows); y++ {
		for x := 0; x < int(cols); x++ {
			out += ufmt.Sprintf(`<rect x="%d" y="%d" width="40" height="40" fill="%s"/>`, x*40, y*40, tableColor[y][x].rgb)
		}
	}
	out += "</svg>"
	return out
}

func Submit(x, y uint64, colorName string) {
	user, ok := retrieveMember(std.PreviousRealm().Address())
	if !ok {
		RegisterMember()
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
	Users.Set(std.PreviousRealm().Address().String(), user)
}
