package metamodel

import (
	"encoding/base64"

	"gno.land/p/demo/ufmt"
)

// sprint is a helper function to convert values to string.
func sprint(v interface{}) string {
	return ufmt.Sprintf("%v", v)
}

// TokenType is a type alias for int64. It is used to represent the number of tokens in a place.
type TokenType = int64

// Place represents a place (circle) in the Petri net.
type Place struct {
	Label    string    // Label of the place.
	Offset   int       // Offset of the place in the Petri net. (used for ordering)
	Initial  TokenType // Initial number of tokens in the place
	Capacity TokenType // Maximum capacity of the place 0=Unlimited
	X        int       // X coordinate of the place
	Y        int       // Y coordinate of the place
}

// Transition represents a transition (rectangle) in the Petri net.
type Transition struct {
	Label  string // Label of the transition.
	Offset int    // Offset of the transition in the Petri net. (used for ordering)
	Role   string // Role of the transition.
	X      int    // X coordinate of the transition
	Y      int    // Y coordinate of the transition
}

// Arrow represents an arrow (arc) in the Petri net.
type Arrow struct {
	Source  string    // Source place or transition
	Target  string    // Target place or transition
	Weight  TokenType // Weight of the arrow
	Consume bool      // Consume bool indicates if the arrow consumes tokens
	Produce bool      // Produce bool indicates if the arrow produces tokens
	Inhibit bool      // Inhibit bool indicates if the arrow inhibits the transition
}

// Pflow represents a Petri net model.
type Pflow struct {
	Places      map[string]Place      // Map of places in the Petri net
	Transitions map[string]Transition // Map of transitions in the Petri net
	Arrows      []Arrow               // List of arrows (Arcs) in the Petri net
}

// StateMachine represents a state machine for the Petri net.
type StateMachine struct {
	Model *Pflow               // Petri net model
	State map[string]TokenType // Current state of the Petri net
}

// NewStateMachine creates a new state machine for the given Petri net model.
func NewStateMachine(m *Pflow) StateMachine {
	return StateMachine{Model: m, State: initialState(m)}
}

// initialState initializes the state of the Petri net model.
func initialState(m *Pflow) map[string]TokenType {
	state := make(map[string]TokenType)
	for label, place := range m.Places {
		state[label] = place.Initial
	}
	return state
}

// Transform applies an action to the state machine and returns true if successful.
func (sm *StateMachine) Transform(action string, multiple ...TokenType) bool {
	m := TokenType(1)
	if len(multiple) == 1 {
		m = multiple[0]
	}
	newState, ok := Execute(sm.State, sm.Model, action, m)
	if !ok {
		return false
	}
	for k, v := range newState {
		sm.State[k] = v
	}
	return true
}

// Execute applies an action to the given state and returns the new state and a boolean indicating success.
func Execute(state map[string]TokenType, model *Pflow, action string, multiple TokenType) (map[string]TokenType, bool) {
	if multiple <= 0 {
		return state, false
	}
	_, ok := model.Transitions[action]
	if !ok {
		return state, false
	}
	newState := make(map[string]TokenType)
	for k, v := range state {
		newState[k] = v
	}
	for _, arc := range model.Arrows {
		if arc.Target == action {
			newState[arc.Source] -= arc.Weight * multiple
			if newState[arc.Source] < 0 { // check for underflow
				return state, false
			}
		} else if arc.Source == action {
			newState[arc.Target] += arc.Weight * multiple
		}
	}
	// check for overflow
	for _, place := range model.Places {
		if newState[place.Label] > place.Capacity {
			return state, false
		}
	}
	return newState, true
}

// NewPflow creates a new Petri net model.
func NewPflow() *Pflow {
	return &Pflow{Places: make(map[string]Place), Transitions: make(map[string]Transition), Arrows: []Arrow{}}
}

// Place adds a place to the Petri net model.
func (net *Pflow) Place(label string, initial TokenType, capacity TokenType, x, y int) {
	net.Places[label] = Place{Label: label, Offset: len(net.Places), Initial: initial, Capacity: capacity, X: x, Y: y}
}

// Transition adds a transition to the Petri net model.
func (net *Pflow) Transition(label string, role string, x, y int) {
	net.Transitions[label] = Transition{Label: label, Offset: len(net.Transitions), Role: role, X: x, Y: y}
}

// arrowPrams checks if the source and target are valid for an arrow.
func (net *Pflow) arrowPrams(source, target string) (consume, produce bool) {
	_, sourceIsPlace := net.Places[source]
	_, targetIsPlace := net.Places[target]
	_, sourceIsTransition := net.Transitions[source]
	_, targetIsTransition := net.Transitions[target]

	if sourceIsPlace && sourceIsTransition {
		panic("source cannot be both a place and a transition")
	}

	if targetIsPlace && targetIsTransition {
		panic("target cannot be both a place and a transition")
	}

	if sourceIsPlace && targetIsTransition {
		consume = true
	}
	if sourceIsTransition && targetIsPlace {
		produce = true
	}
	return
}

// Add methods to add places, transitions, and arrows to the Petri net model.
func (net *Pflow) Arrow(source, target string, weight TokenType) {
	consume, produce := net.arrowPrams(source, target)
	net.Arrows = append(net.Arrows, Arrow{Source: source, Target: target, Weight: weight, Consume: consume, Produce: produce, Inhibit: false})
}

// Guard adds a guard to the Petri net model.
func (net *Pflow) Guard(source, target string, weight TokenType) {
	consume, produce := net.arrowPrams(source, target)
	net.Arrows = append(net.Arrows, Arrow{Source: source, Target: target, Weight: weight, Consume: consume, Produce: produce, Inhibit: true})
}

// Display represents the Svg display of the Petri net model.
type Display struct {
	Buffer string
	Model  *Pflow
	State  map[string]TokenType
}

// NewDisplay creates a new display for the given Petri net model and state.
func NewDisplay(model *Pflow, state ...map[string]TokenType) Display {
	if len(state) == 1 {
		return Display{Model: model, State: state[0]}
	}
	return Display{Model: model, State: make(map[string]TokenType)}
}

// GetViewPort calculates the viewport dimensions for the Petri net model.
func (d *Display) GetViewPort() (x1 int, y1 int, width int, height int) {
	var minX int = 0
	var minY int = 0
	var limitX int = 0
	var limitY int = 0

	for _, p := range d.Model.Places {
		if limitX < p.X {
			limitX = p.X
		}
		if limitY < p.Y {
			limitY = p.Y
		}
		if minX == 0 || minX > p.X {
			minX = p.X
		}
		if minY == 0 || minY > p.Y {
			minY = p.Y
		}
	}
	for _, t := range d.Model.Transitions {
		if limitX < t.X {
			limitX = t.X
		}
		if limitY < t.Y {
			limitY = t.Y
		}
		if minX == 0 || minX > t.X {
			minX = t.X
		}
		if minY == 0 || minY > t.Y {
			minY = t.Y
		}
	}
	const margin = 60
	x1 = minX - margin
	y1 = minY - margin
	x2 := limitX + margin
	y2 := limitY + margin

	return x1, y1, x2 - x1, y2 - y1
}

// NewSvgImage creates a new Svg image for the Petri net model
func (d *Display) NewSvgImage() {
	x1, y1, width, height := d.GetViewPort()
	d.Buffer += ufmt.Sprintf("<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"%v\" height=\"%v\" viewBox=\"%v %v %v %v\">", width, height, x1, y1, width, height)
	d.Rect(0, 0, width, height, "fill=\"#ffffff\"")
	d.WriteDefs()
}

// WriteDefs writes the Svg definitions for the display.
func (d *Display) WriteDefs() {
	d.Buffer += "<defs>" +
		"<marker id=\"markerArrow1\" markerWidth=\"23\" markerHeight=\"13\" refX=\"31\" refY=\"6\" orient=\"auto\">" +
		"<rect width=\"28\" height=\"3\" fill=\"white\" stroke=\"white\" x=\"3\" y=\"5\"/>" +
		"<path d=\"M2,2 L2,11 L10,6 L2,2\"/>" +
		"</marker>" +
		"<marker id=\"markerInhibit1\" markerWidth=\"23\" markerHeight=\"13\" refX=\"31\" refY=\"6\" orient=\"auto\">" +
		"<rect width=\"28\" height=\"3\" fill=\"white\" stroke=\"white\" x=\"3\" y=\"5\"/>" +
		"<circle cx=\"5\" cy=\"6.5\" r=\"4\"/>" +
		"</marker>" +
		"</defs>"
}

// Group starts a new group in the Svg.
func (d *Display) Gend() {
	d.WriteElement("</g>")
}

// WriteElement writes an element to the Svg buffer.
func (d *Display) WriteElement(element string) {
	d.Buffer += element
}

// Render renders the Petri net model to Svg.
func (d *Display) Render() {
	for _, arc := range d.Model.Arrows {
		d.ArcElement(arc)
	}
	for label, place := range d.Model.Places {
		d.PlaceElement(label, place)
	}
	for label, transition := range d.Model.Transitions {
		d.TransitionElement(label, transition)
	}
	d.EndSvg()
}

// PlaceElement renders a place element in the Svg.
func (d *Display) PlaceElement(label string, place Place) {
	d.Group()
	d.Circle(place.X, place.Y, 16, "stroke-width=\"1.5\" fill=\"#ffffff\" stroke=\"#000000\"")
	d.Text(place.X-18, place.Y-20, label, "font-size=\"small\"")
	x := place.X
	y := place.Y
	tokens := place.Initial
	if state, ok := d.State[label]; ok {
		tokens = state
	}
	if tokens > 0 {
		if tokens == 1 {
			d.Circle(x, y, 2, "fill=\"#000000\" stroke=\"#000000\"")
		} else if tokens < 10 {
			d.Text(x-4, y+5, sprint(tokens), "font-size=\"large\"")
		} else {
			d.Text(x-7, y+5, sprint(tokens), "font-size=\"small\"")
		}
	}
	d.Gend()
}

// ArcElement renders an arc element in the Svg.
func (d *Display) ArcElement(arc Arrow) {
	d.Group()
	marker := "url(#markerArrow1)"
	if arc.Inhibit {
		marker = "url(#markerInhibit1)"
	}
	extra := "stroke=\"#000000\" fill=\"#000000\" marker-end=\"" + marker + "\""

	var p Place
	var t Transition
	if arc.Inhibit {
		if place, ok := d.Model.Places[arc.Source]; ok {
			p = place
			t = d.Model.Transitions[arc.Target]
		} else {
			p = d.Model.Places[arc.Target]
			t = d.Model.Transitions[arc.Source]
		}
	} else {
		if place, ok := d.Model.Places[arc.Source]; ok {
			p = place
			t = d.Model.Transitions[arc.Target]
		} else {
			p = d.Model.Places[arc.Target]
			t = d.Model.Transitions[arc.Source]
		}
	}

	if place, ok := d.Model.Places[arc.Source]; ok {
		p = place
		t = d.Model.Transitions[arc.Target]
		d.Line(p.X, p.Y, t.X, t.Y, extra)
		midX := (p.X + t.X) / 2
		midY := (p.Y+t.Y)/2 - 8
		weight := arc.Weight
		d.Text(midX-4, midY+4, sprint(weight), "font-size=\"small\"")
	} else {
		p = d.Model.Places[arc.Target]
		t = d.Model.Transitions[arc.Source]
		d.Line(t.X, t.Y, p.X, p.Y, extra)
		midX := (t.X + p.X) / 2
		midY := (t.Y+p.Y)/2 - 8
		weight := arc.Weight
		d.Text(midX-4, midY+4, sprint(weight), "font-size=\"small\"")
	}
	d.Gend()
}

// TransitionElement renders a transition element in the Svg.
func (d *Display) TransitionElement(label string, transition Transition) {
	d.Group()
	x := transition.X - 17
	y := transition.Y - 17
	d.Rect(x, y, 30, 30, "stroke=\"#000000\" fill=\"#ffffff\" rx=\"4\"")
	d.Text(x, y-8, label, "font-size=\"small\"")
	d.Gend()
}

// EndSvg ends the Svg image.
func (d *Display) EndSvg() {
	d.Buffer += "</svg>"
}

// ToSvg returns the Svg representation of the Petri net model.
func (net *Pflow) ToSvg(state ...map[string]TokenType) string {
	s := make(map[string]TokenType)
	if len(state) == 1 {
		s = state[0]
	}
	d := NewDisplay(net, s)
	d.NewSvgImage()
	d.Render()
	return d.Buffer
}

// Rect draws a rectangle in the Svg.
func (d *Display) Rect(x, y, width, height int, extra string) {
	d.WriteElement("<rect x=\"" + sprint(x) + "\" y=\"" + sprint(y) + "\" width=\"" + sprint(width) + "\" height=\"" + sprint(height) + "\" " + extra + " />")
}

// Circle draws a circle in the Svg.
func (d *Display) Circle(x, y, radius int, extra string) {
	d.WriteElement("<circle cx=\"" + sprint(x) + "\" cy=\"" + sprint(y) + "\" r=\"" + sprint(radius) + "\" " + extra + " />")
}

// Text draws text in the Svg.
func (d *Display) Text(x, y int, txt, extra string) {
	d.WriteElement("<text x=\"" + sprint(x) + "\" y=\"" + sprint(y) + "\" " + extra + ">" + txt + "</text>")
}

// Line draws a line in the Svg.
func (d *Display) Line(x1, y1, x2, y2 int, extra string) {
	d.WriteElement("<line x1=\"" + sprint(x1) + "\" y1=\"" + sprint(y1) + "\" x2=\"" + sprint(x2) + "\" y2=\"" + sprint(y2) + "\" " + extra + " />")
}

// Group starts a new group in the Svg.
func (d *Display) Group() {
	d.WriteElement("<g>")
}

// ToJson returns a JSON representation of the Petri net model.
// Open the JSON with https://pflow.xyz/editor to interact with the state machine.
func (net *Pflow) ToJson() string {
	jsonStr := "{"
	jsonStr += "\"modelType\":\"petriNet\","
	jsonStr += "\"version\":\"v0\","
	jsonStr += "\"places\":{"
	first := true
	for label, place := range net.Places {
		if !first {
			jsonStr += ","
		}
		first = false
		jsonStr += "\"" + label + "\":{"
		jsonStr += "\"offset\":" + sprint(place.Offset) + ","
		jsonStr += "\"initial\":" + sprint(place.Initial) + ","
		jsonStr += "\"capacity\":" + sprint(place.Capacity) + ","
		jsonStr += "\"x\":" + sprint(place.X) + ","
		jsonStr += "\"y\":" + sprint(place.Y)
		jsonStr += "}"
	}
	jsonStr += "},"
	jsonStr += "\"transitions\":{"
	first = true
	for label, transition := range net.Transitions {
		if !first {
			jsonStr += ","
		}
		first = false
		jsonStr += "\"" + label + "\":{"
		jsonStr += "\"x\":" + sprint(transition.X) + ","
		jsonStr += "\"y\":" + sprint(transition.Y)
		jsonStr += "}"
	}
	jsonStr += "},"
	jsonStr += "\"arcs\":["
	first = true
	for _, arc := range net.Arrows {
		if !first {
			jsonStr += ","
		}
		first = false
		jsonStr += "{"
		jsonStr += "\"source\":\"" + arc.Source + "\","
		jsonStr += "\"target\":\"" + arc.Target + "\""
		if arc.Weight != 0 {
			jsonStr += ",\"weight\":" + sprint(arc.Weight)
		}
		if arc.Inhibit {
			jsonStr += ",\"inhibit\":true"
		}
		jsonStr += "}"
	}
	jsonStr += "]"
	jsonStr += "}"
	return jsonStr
}

// ToBase64 returns a base64 encoded JSON representation of the Petri net model.
// Open the base64 encoded JSON with https://pflow.xyz/editor to interact with the state machine.
func (net *Pflow) ToBase64() string {
	return base64.StdEncoding.EncodeToString([]byte(net.ToJson()))
}

// ToJsonDataUrl returns a data url for the JSON representation of the Petri net model.
// Open the data url in a browser to interact with the state machine.
func (net *Pflow) ToJsonDataUrl() string {
	return "data:text/json;base64," + net.ToBase64()
}

// ToSvgDataUrl returns a data url for the Svg representation of the Petri net model.
// Open the data url in a browser to view the state machine.
func (net *Pflow) ToSvgDataUrl() string {
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(net.ToSvg()))
}

// ToImageMarkdown returns a markdown image of the Petri net model.
// Open the image in a browser to view the state machine.
// The image is a base64 encoded SVG.
// The image is not interactive.
func (net *Pflow) ToImageMarkdown() string {
	return "![Petri-net](" + net.ToSvgDataUrl() + ")"
}

// ToLinkMarkdown returns a markdown link to the Petri net model.
// The urlBase is the base url of the pflow server. Default is https://pflow.dev
// The link is to the pflow editor with the Petri net model as a base64 encoded query parameter.
// Open the link in a browser to interact with the state machine.
func (net *Pflow) ToLinkMarkdown(urlBase ...string) string {
	url := "https://pflow.dev"
	if len(urlBase) > 0 {
		url = urlBase[0]
	}
	data := net.ToBase64()
	return "[![Petri-net](" + url + "/img/?b=" + data + ")](" + url + "/?b=" + data + ")"
}
