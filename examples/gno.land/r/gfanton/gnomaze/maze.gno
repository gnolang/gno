// This package demonstrate the capability of gno to build dynamic svg image
// based on different query parameters.
// Raycasting implementation as been heavily inspired by this project: https://github.com/AZHenley/raycasting

package gnomaze

import (
	"chain/runtime"
	"encoding/base64"
	"hash/adler32"
	"math"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gno.land/p/moul/txlink"
	"gno.land/p/nt/ufmt"
	"gno.land/r/leon/hor"
)

const baseLevel = 7

// Constants for cell dimensions
const (
	cellSize = 1.0
	halfCell = cellSize / 2
)

type CellKind int

const (
	CellKindEmpty = iota
	CellKindWall
)

var (
	level            int = 1
	salt             int64
	maze             [][]int
	endPos, startPos Position
)

func init() {
	// Generate the map
	seed := uint64(runtime.ChainHeight())
	rng := rand.New(rand.NewPCG(seed, uint64(time.Now().Unix())))
	generateLevel(rng, level)
	salt = rng.Int64()

	// Register to hor
	hor.Register(cross, "GnoMaze, A 3D Maze Game", "")
}

// Position represents the X, Y coordinates in the maze
type Position struct{ X, Y int }

// Player represents a player with position and viewing angle
type Player struct {
	X, Y, Angle, FOV float64
}

// PlayerState holds the player's grid position and direction
type PlayerState struct {
	CellX     int // Grid X position
	CellY     int // Grid Y position
	Direction int // 0-7 (0 = east, 1 = SE, 2 = S, etc.)
}

// Angle calculates the direction angle in radians
func (p *PlayerState) Angle() float64 {
	return float64(p.Direction) * math.Pi / 4
}

// Position returns the player's exact position in the grid
func (p *PlayerState) Position() (float64, float64) {
	return float64(p.CellX) + halfCell, float64(p.CellY) + halfCell
}

// SumCode returns a hash string based on the player's position
func (p *PlayerState) SumCode() string {
	a := adler32.New()

	var width int
	if len(maze) > 0 {
		width = len(maze[0])
	}

	ufmt.Fprintf(a, "%d-%d-%d", p.CellY*width+p.CellX, level, salt)
	return strconv.FormatUint(uint64(a.Sum32()), 10)
}

// Move updates the player's position based on movement deltas
func (p *PlayerState) Move(dx, dy int) {
	newX := p.CellX + dx
	newY := p.CellY + dy

	if newY >= 0 && newY < len(maze) && newX >= 0 && newX < len(maze[0]) {
		if maze[newY][newX] == 0 {
			p.CellX = newX
			p.CellY = newY
		}
	}
}

// Rotate changes the player's direction
func (p *PlayerState) Rotate(clockwise bool) {
	if clockwise {
		p.Direction = (p.Direction + 1) % 8
	} else {
		p.Direction = (p.Direction + 7) % 8
	}
}

// GenerateNextLevel validates the answer and generates a new level
func GenerateNextLevel(cur realm, answer string) {
	seed := uint64(runtime.ChainHeight())
	rng := rand.New(rand.NewPCG(seed, uint64(time.Now().Unix())))

	endState := PlayerState{CellX: endPos.X, CellY: endPos.Y}
	hash := endState.SumCode()
	if hash != answer {
		panic("invalid answer")
	}

	// Generate new map
	level++
	salt = rng.Int64()
	generateLevel(rng, level)
}

// generateLevel creates a new maze for the given level
func generateLevel(rng *rand.Rand, level int) {
	if level < 0 {
		panic("invalid level")
	}

	size := level + baseLevel
	maze, startPos, endPos = generateMap(rng, size, size)
}

// generateMap creates a random maze using a depth-first search algorithm.
func generateMap(rng *rand.Rand, width, height int) ([][]int, Position, Position) {
	// Initialize the maze grid filled with walls.
	m := make([][]int, height)
	for y := range m {
		m[y] = make([]int, width)
		for x := range m[y] {
			m[y][x] = CellKindWall
		}
	}

	// Define start position and initialize stack for DFS
	start := Position{1, 1}
	stack := []Position{start}
	m[start.Y][start.X] = CellKindEmpty

	// Initialize distance matrix and track farthest
	dist := make([][]int, height)
	for y := range dist {
		dist[y] = make([]int, width)
		for x := range dist[y] {
			dist[y][x] = -1
		}
	}
	dist[start.Y][start.X] = CellKindEmpty
	maxDist := 0
	candidates := []Position{start}

	// Possible directions for movement: right, left, down, up
	directions := []Position{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}

	// Generate maze paths using DFS
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		var dirCandidates []struct {
			next, wall Position
		}

		// Evaluate possible candidates for maze paths
		for _, d := range directions {
			nx, ny := current.X+d.X*2, current.Y+d.Y*2
			wx, wy := current.X+d.X, current.Y+d.Y

			// Check if the candidate position is within bounds and still a wall
			if nx > 0 && nx < width-1 && ny > 0 && ny < height-1 && m[ny][nx] == 1 {
				dirCandidates = append(dirCandidates, struct{ next, wall Position }{
					Position{nx, ny}, Position{wx, wy},
				})
			}
		}

		// If candidates are available, choose one and update the maze
		if len(dirCandidates) > 0 {
			chosen := dirCandidates[rng.IntN(len(dirCandidates))]
			m[chosen.wall.Y][chosen.wall.X] = CellKindEmpty
			m[chosen.next.Y][chosen.next.X] = CellKindEmpty

			// Update distance for the next cell
			currentDist := dist[current.Y][current.X]
			nextDist := currentDist + 2
			dist[chosen.next.Y][chosen.next.X] = nextDist

			// Update maxDist and candidates
			if nextDist > maxDist {
				maxDist = nextDist
				candidates = []Position{chosen.next}
			} else if nextDist == maxDist {
				candidates = append(candidates, chosen.next)
			}

			stack = append(stack, current, chosen.next)
		}
	}

	// Select a random farthest position as the end
	var end Position
	if len(candidates) > 0 {
		end = candidates[rng.IntN(len(candidates))]
	} else {
		end = Position{width - 2, height - 2} // Fallback to bottom-right
	}

	return m, start, end
}

// castRay simulates a ray casting in the maze to find walls
func castRay(playerX, playerY, rayAngle float64, m [][]int) (distance float64, wallHeight float64, endCellHit bool, endDistance float64) {
	x, y := playerX, playerY
	dx, dy := math.Cos(rayAngle), math.Sin(rayAngle)
	steps := 0
	endCellHit = false
	endDistance = 0.0

	for {
		ix, iy := int(math.Floor(x)), int(math.Floor(y))
		if ix == endPos.X && iy == endPos.Y {
			endCellHit = true
			endDistance = math.Sqrt(math.Pow(x-playerX, 2) + math.Pow(y-playerY, 2))
		}

		if iy < 0 || iy >= len(m) || ix < 0 || ix >= len(m[0]) || m[iy][ix] != 0 {
			break
		}

		x += dx * 0.1
		y += dy * 0.1
		steps++
		if steps > 400 {
			break
		}
	}

	distance = math.Sqrt(math.Pow(x-playerX, 2) + math.Pow(y-playerY, 2))
	wallHeight = 300.0 / distance
	return
}

// GenerateSVG creates an SVG representation of the maze scene
func GenerateSVG(cur realm, p *PlayerState) string {
	const (
		svgWidth, svgHeight = 800, 600
		offsetX, offsetY    = 0.0, 500.0
		groundLevel         = 300
		rays                = 124
		fov                 = math.Pi / 4
		miniMapSize         = 100.0
		visibleCells        = 7
		dirLen              = 2.0
	)

	m := maze
	playerX, playerY := p.Position()
	angle := p.Angle()

	sliceWidth := float64(svgWidth) / float64(rays)
	angleStep := fov / float64(rays)

	var svg strings.Builder
	svg.WriteString(`<svg width="800" height="600" xmlns="http://www.w3.org/2000/svg">`)
	svg.WriteString(`<rect x="0" y="0" width="800" height="300" fill="rgb(20,40,20)"/>`)
	svg.WriteString(`<rect x="0" y="300" width="800" height="300" fill="rgb(40,60,40)"/>`)

	var drawBanana func()
	for i := 0; i < rays; i++ {
		rayAngle := angle - fov/2 + float64(i)*angleStep
		distance, wallHeight, endHit, endDist := castRay(playerX, playerY, rayAngle, m)
		darkness := 1.0 + distance/4.0
		colorVal1 := int(180.0 / darkness)
		colorVal2 := int(32.0 / darkness)
		yPos := groundLevel - wallHeight/2

		ufmt.Fprintf(&svg,
			`<rect x="%f" y="%f" width="%f" height="%f" fill="rgb(%d,69,%d)"/>`,
			float64(i)*sliceWidth, yPos, sliceWidth, wallHeight, colorVal1, colorVal2)

		if drawBanana != nil {
			continue // Banana already drawn
		}

		// Only draw banana if the middle ray hit the end
		// XXX: improve this by checking for a hit in the middle of the end cell
		if i == rays/2 && endHit && endDist < distance {
			iconHeight := 10.0 / endDist
			scale := iconHeight / 100
			x := float64(i)*sliceWidth + sliceWidth/2
			y := groundLevel + 20 + (iconHeight*scale)/2

			drawBanana = func() {
				ufmt.Fprintf(&svg,
					`<g transform="translate(%f %f) scale(%f)">%s</g>`,
					x, y, scale, string(svgassets["banana"]),
				)
			}
		}
	}

	if drawBanana != nil {
		drawBanana()
	}

	playerCellX, playerCellY := int(math.Floor(playerX)), int(math.Floor(playerY))

	xStart := max(0, playerCellX-visibleCells/2)
	xEnd := min(len(m[0]), playerCellX+visibleCells/2+1)

	yStart := max(0, playerCellY-visibleCells/2)
	yEnd := min(len(m), playerCellY+visibleCells/2+1)

	scaleX := miniMapSize / float64(xEnd-xStart)
	scaleY := miniMapSize / float64(yEnd-yStart)

	for y := yStart; y < yEnd; y++ {
		for x := xStart; x < xEnd; x++ {
			color := "black"
			if m[y][x] == 1 {
				color = "rgb(149,0,32)"
			}
			ufmt.Fprintf(&svg,
				`<rect x="%f" y="%f" width="%f" height="%f" fill="%s"/>`,
				float64(x-xStart)*scaleX+offsetX, float64(y-yStart)*scaleY+offsetY, scaleX, scaleY, color)
		}
	}

	px := (playerX-float64(xStart))*scaleX + offsetX
	py := (playerY-float64(yStart))*scaleY + offsetY
	ufmt.Fprintf(&svg, `<circle cx="%f" cy="%f" r="%f" fill="rgb(200,200,200)"/>`, px, py, scaleX/2)

	dx := math.Cos(angle) * dirLen
	dy := math.Sin(angle) * dirLen
	ufmt.Fprintf(&svg,
		`<line x1="%f" y1="%f" x2="%f" y2="%f" stroke="rgb(200,200,200)" stroke-width="1"/>`,
		px, py, (playerX+dx-float64(xStart))*scaleX+offsetX, (playerY+dy-float64(yStart))*scaleY+offsetY)

	svg.WriteString(`</svg>`)
	return svg.String()
}

// renderGrid3D creates a 3D view of the grid
func renderGrid3D(cur realm, p *PlayerState) string {
	svg := GenerateSVG(cur, p)
	base64SVG := base64.StdEncoding.EncodeToString([]byte(svg))
	return ufmt.Sprintf("![SVG Image](data:image/svg+xml;base64,%s)", base64SVG)
}

// generateDirLink generates a link to change player direction
func generateDirLink(path string, p *PlayerState, action string) string {
	newState := *p // Make copy

	switch action {
	case "forward":
		dx, dy := directionDeltas(newState.Direction)
		newState.Move(dx, dy)
	case "left":
		newState.Rotate(false)
	case "right":
		newState.Rotate(true)
	}

	vals := make(url.Values)
	vals.Set("x", strconv.Itoa(newState.CellX))
	vals.Set("y", strconv.Itoa(newState.CellY))
	vals.Set("dir", strconv.Itoa(newState.Direction))

	vals.Set("sum", newState.SumCode())
	return path + "?" + vals.Encode()
}

// isPlayerTouchingWall checks if the player's position is inside a wall
func isPlayerTouchingWall(x, y float64) bool {
	ix, iy := int(math.Floor(x)), int(math.Floor(y))
	if iy < 0 || iy >= len(maze) || ix < 0 || ix >= len(maze[0]) {
		return true
	}
	return maze[iy][ix] == CellKindEmpty
}

// directionDeltas provides deltas for movement based on direction
func directionDeltas(d int) (x, y int) {
	s := []struct{ x, y int }{
		{1, 0},   // 0 == E
		{1, 1},   // SE
		{0, 1},   // S
		{-1, 1},  // SW
		{-1, 0},  // W
		{-1, -1}, // NW
		{0, -1},  // N
		{1, -1},  // NE
	}[d]
	return s.x, s.y
}

// atoiDefault converts string to integer with a default fallback
func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	i, _ := strconv.Atoi(s)
	return i
}

// Render renders the game interface
func Render(path string) string {
	u, _ := url.Parse(path)
	query := u.Query()

	p := PlayerState{
		CellX:     atoiDefault(query.Get("x"), startPos.X),
		CellY:     atoiDefault(query.Get("y"), startPos.Y),
		Direction: atoiDefault(query.Get("dir"), 0), // Start facing east
	}

	cpath := strings.TrimPrefix(runtime.CurrentRealm().PkgPath(), runtime.ChainDomain())
	psum := p.SumCode()
	reset := "[reset](" + cpath + ")"

	if startPos.X != p.CellX || startPos.Y != p.CellY {
		if sum := query.Get("sum"); psum != sum {
			return "invalid sum : " + reset
		}
	}

	if endPos.X == p.CellX && endPos.Y == p.CellY {
		return strings.Join([]string{
			ufmt.Sprintf("### Congrats you win level %d !!", level),
			ufmt.Sprintf("Code for next level is: %s", psum),
			ufmt.Sprintf("[Generate Next Level: %d](%s)", level+1, txlink.Call("GenerateNextLevel", "answer", psum)),
		}, "\n\n")
	}

	// Generate commands
	commands := strings.Join([]string{
		"<gno-columns>",
		"|||",
		ufmt.Sprintf("[▲](%s)", generateDirLink(cpath, &p, "forward")),
		"|||",
		"</gno-columns>",
		"<gno-columns>",
		ufmt.Sprintf("[◄](%s)", generateDirLink(cpath, &p, "left")),
		"|||",
		"|||",
		ufmt.Sprintf("[►](%s)", generateDirLink(cpath, &p, "right")),
		"</gno-columns>",
	}, "\n\n")

	// Generate view
	view := strings.Join([]string{
		"<gno-columns>",
		renderGrid3D(cross, &p),
		"</gno-columns>",
	}, "\n\n")

	return strings.Join([]string{
		"## Find the banana: Level " + strconv.Itoa(level),
		"---", view, "---", commands, "---",
		reset,
		ufmt.Sprintf("Position: (%d, %d) Direction: %fπ", p.CellX, p.CellY, float64(p.Direction)/math.Pi),
	}, "\n\n")
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
