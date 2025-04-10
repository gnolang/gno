package bmp

import (
	"gno.land/p/demo/ufmt"
	"net/url"
	"strings"
)

var (
	pixels    = NewGrid(50)
	published = false
	bounds    = Bounds{
		X1: 0,
		Y1: 13,
		X2: 40,
		Y2: 42,
	}
)

type Bounds struct {
	X1 int
	Y1 int
	X2 int
	Y2 int
}

type ColorPixel struct {
	X int
	Y int
	R uint8
	G uint8
	B uint8
}

// REVIEW: tuned the bounds to 10-40 and 15-45 to constraint the image
// 30x30 seems to be about the largest size we can render in SVG without hitting the limit

func SetPixel(x, y, r, g, b int) {
	if published {
		panic("Cannot set pixel after publishing")
	}
	if x < bounds.X1 || x > bounds.X2 || y < bounds.Y1 || y > bounds.Y2 {
		return
	}
	if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 {
		return
	}
	pixels[x*50+y] = ColorPixel{
		X: x,
		Y: y,
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
	}
}

func Publish() {
	published = true
}

func Render(path string) string {
	if strings.HasPrefix(path, "gno") {
		return "```\n" + Source() + "\n```"
	}
	svg := GenerateSVG()
	escaped := url.PathEscape(svg)
	return "![bmp](data:image/svg+xml;utf8," + escaped + ")"
}

func GenerateSVG() string {
	tileSize := 1

	var sb strings.Builder

	// Compute viewbox
	var maxX, maxY int
	var minX, minY int
	for _, p := range pixels {
		if p.R == 0 && p.G == 0 && p.B == 0 {
			continue // Skip black pixels
		}
		if minX == 0 && minY == 0 {
			minX = p.X
			minY = p.Y
		}
		if p.X < minX {
			minX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	minX = minX * tileSize
	minY = minY * tileSize
	width := (maxX - minX + 1) * tileSize
	height := (maxY - minY + 1) * tileSize

	sb.WriteString(ufmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="400" height="400" viewBox="%d %d %d %d">`,
		minX, minY, width, height))

	sb.WriteString(`
  <defs>
    <rect id="p" width="1" height="1" rx="0.2" ry="0.2"/>
  </defs>
  <style>
`)

	colorClassMap := make(map[string]string)
	classCount := 0
	for _, p := range pixels {
		key := ufmt.Sprintf("%d_%d_%d", p.R, p.G, p.B)
		if _, ok := colorClassMap[key]; !ok {
			className := ufmt.Sprintf("c%d", classCount)
			colorClassMap[key] = className
			classCount++
			sb.WriteString(ufmt.Sprintf("  .%s { fill: rgb(%d,%d,%d); }\n", className, p.R, p.G, p.B))
		}
	}

	sb.WriteString("  </style>\n")

	sb.WriteString("<g id=\"bmp-image\">\n")
	for _, p := range pixels {
		if p.R == 0 && p.G == 0 && p.B == 0 {
			continue // Skip black pixels
		}
		key := ufmt.Sprintf("%d_%d_%d", p.R, p.G, p.B)
		className := colorClassMap[key]
		sb.WriteString(ufmt.Sprintf(
			`  <use href="#p" x="%d" y="%d" class="%s"/>`+"\n",
			p.X*tileSize, p.Y*tileSize, className))
	}
	sb.WriteString("</g>\n")

	sb.WriteString(`</svg>`)

	return sb.String()
}

func NewGrid(size ...int) []ColorPixel {
	gridSize := 50
	if len(size) > 0 {
		gridSize = size[0]
	}
	pixels := make([][]ColorPixel, gridSize)
	for i := range pixels {
		pixels[i] = make([]ColorPixel, gridSize)
		for j := range pixels[i] {
			pixels[i][j] = ColorPixel{}
		}
	}

	// Flatten the 2D slice into a 1D slice
	var flatPixels []ColorPixel
	for _, row := range pixels {
		flatPixels = append(flatPixels, row...)
	}

	return flatPixels
}

const heading = `
import (
    image "gno.land/r/stackdump/bmp"
)

func init() {
    p := image.SetPixel()
`

const footer = `
    image.Publish() // finalize the image
}
`

func Source() string {
	src := heading
	livePixels := 0
	for i := 0; i < len(pixels); i++ {
		p := pixels[i]
		if p.R == 0 && p.G == 0 && p.B == 0 {
			continue // Skip black pixels
		}
		src += ufmt.Sprintf("\tp(%d, %d, %d, %d, %d)\n", p.X, p.Y, p.R, p.G, p.B)
		livePixels++
	}
	src += ufmt.Sprintf("\t// Found %d/%d live Pixels", livePixels, len(pixels)) + footer
	return src
}
