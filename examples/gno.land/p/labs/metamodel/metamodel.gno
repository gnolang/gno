package metamodel

import (
	"math"
	"net/url"
	"strings"

	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"

	"sort"

	"gno.land/p/demo/ufmt"
)

// TokenType represents multiple token types as a slice of integers.
type TokenType []int64 // REVIEW: should this be int256

func (tt TokenType) String() string {
	if len(tt) == 0 {
		return "[]"
	}
	var sb strings.Builder
	sb.WriteString("[")
	for i, v := range tt {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(ufmt.Sprintf("%d", v))
	}
	sb.WriteString("]")
	return sb.String()
}

// T converts a variadic list of values into a TokenType.
func T(t ...any) TokenType {
	ts := make(TokenType, len(t))
	for i, v := range t {
		switch v := v.(type) {
		case int:
			ts[i] = int64(v)
		case int64:
			ts[i] = v
		case string:
			ts[i] = 1
		default:
			ts[i] = -1 // Default to -1 for unsupported types.
		}
	}
	return ts
}

// Place represents a place in the Petri net.
type Place struct {
	Label    string    // Label of the place.
	Offset   int       // Offset of the place in the Petri net.
	Tokens   TokenType // Current number of tokens for each type.
	Initial  TokenType // Initial number of tokens for each type.
	Capacity TokenType // Maximum capacity for each token type (0 = unlimited).
	X        int       // X coordinate of the place.
	Y        int       // Y coordinate of the place.
	Binding  any       // Optional binding for the place.
}

// Arc represents an arc in the Petri net.
type Arc struct {
	Source  string // Source place or transition.
	Target  string // Target place or transition.
	Weight  int64  // Weight for the token type.
	Inhibit bool   // Indicates if the arc inhibits the transition.
	Object  string // Optional object name for the arrow.
}

// Arrow represents an arrow in the Petri net.
type Arrow struct {
	Source  string    // Source place or transition.
	Target  string    // Target place or transition.
	Weight  TokenType // Weight for each token type used in the model.
	Inhibit bool      // Indicates if the arc inhibits the transition.
	Binding any       // Optional binding for the arc.
}

// Transition represents a transition in the Petri net.
type Transition struct {
	Label   string  // Label of the transition.
	Offset  int     // Offset of the transition in the Petri net.
	X       int     // X coordinate of the transition.
	Y       int     // Y coordinate of the transition.
	Rate    float64 // Rate of the transition (for continuous models).
	Binding any     // Optional binding for the transition.
}

// fmtWeight formats the weight of a TokenType.
func fmtWeight(weight TokenType) (out string) {
	if len(weight) == 0 || weight[0] == 0 {
		return ""
	}
	if len(weight) == 1 {
		out = ufmt.Sprintf("%d", weight[0])
	} else {
		out = ufmt.Sprintf("%v", weight)
	}
	if weight[0] >= 1_000_000 {
		out = ufmt.Sprintf("%dM", weight[0]/1_000_000)
	} else if weight[0] >= 1000 {
		out = ufmt.Sprintf("%dK", weight[0]/1000)
	} else {
		out = ufmt.Sprintf("%d", weight[0])
	}
	return out
}

// assertValidObjectName ensures that an object name is valid.
func assertValidObjectName(name string) {
	if len(name) == 0 || name[0] != '$' {
		panic("object names must start with $")
	}
	for i, ch := range name[1:] {
		if !(ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9' && i > 0)) {
			panic("object names must be valid identifiers")
		}
	}
}

// Model represents the Petri net model.
type Model struct {
	Objects     []string              // Additional $objects associated with the model.
	Places      map[string]Place      // Map of places.
	Transitions map[string]Transition // Map of transitions.
	Arrows      []Arrow               // List of arcs.
	Binding     any                   // Optional binding for the model.
	RenderOpts  map[string]any        // Rendering options for the model.
}

// l1 := New(m.Objects, m.Places) //.Cid()
// l2 :  New(l1, Arrows{})

func New(inputs ...any) *Model {
	m := &Model{
		Places:      map[string]Place{},
		Transitions: map[string]Transition{},
		Arrows:      []Arrow{},
	}

	for _, input := range inputs {
		switch v := input.(type) {
		case []string:
			for _, obj := range v {
				AssertValidObjectName(obj)
				m.Objects = append(m.Objects, obj)
			}
		default:
			continue
		}
	}

	for _, input := range inputs {
		switch v := input.(type) {
		case *Model:
			// m.Objects = append(m.Objects, v.Objects...) // REVIEW: should we merge objects?
			for label, place := range v.Places {
				m.Places[label] = place
			}
			for label, transition := range v.Transitions {
				m.Transitions[label] = transition
			}
			m.Arrows = append(m.Arrows, v.Arrows...)
		case map[string]Place:
			for label, place := range v {
				m.Places[label] = place
			}
		case map[string]Transition:
			for label, transition := range v {
				m.Transitions[label] = transition
			}
		case []Arrow:
			m.Arrows = append(m.Arrows, v...)
		case map[string]any: // REVIEW: we currently assume a finalized set of objects is passe (i.e. a superset of all objects)
			m.RenderOpts = v
		default:
			continue
		}
	}

	offset := 0
	for label, place := range m.Places {
		place.Offset = offset
		m.Places[label] = place
		offset++
	}

	offset = 0
	for label, transition := range m.Transitions {
		transition.Offset = offset
		m.Transitions[label] = transition
		offset++
	}

	return m
}
func (model *Model) T(val int64, object string) TokenType {
	size := 1
	if len(model.Objects) > 0 {
		size = len(model.Objects)
	}

	wt := make(TokenType, size)
	if object != "" && len(model.Objects) > 0 {
		for i, obj := range model.Objects {
			if obj == object {
				wt[i] = val
				break
			}
		}
	} else {
		wt[0] = val // default to first object
	}
	return wt
}

func (model *Model) Arrow(opts Arc) Arrow {
	if opts.Source == "" {
		panic("Source must be a non-empty string")
	}
	if opts.Target == "" {
		panic("Target must be a non-empty string")
	}
	if opts.Weight < 0 {
		panic("Weight must be a non-negative int64")
	}

	size := 1
	if len(model.Objects) > 0 {
		size = len(model.Objects)
	}

	wt := make(TokenType, size)
	if opts.Object != "" && len(model.Objects) > 0 {
		for i, obj := range model.Objects {
			if obj == opts.Object {
				wt[i] = opts.Weight
				break
			}
		}
	} else {
		wt[0] = opts.Weight
	}

	return Arrow{
		Source:  opts.Source,
		Target:  opts.Target,
		Weight:  wt,
		Inhibit: opts.Inhibit,
	}
}

func (model *Model) CanvasSize() (width, height int) {
	padding := 50
	maxX, maxY := 0, 0
	for _, place := range model.Places {
		if place.X > maxX {
			maxX = place.X
		}
		if place.Y > maxY {
			maxY = place.Y
		}
	}
	for _, transition := range model.Transitions {
		if transition.X > maxX {
			maxX = transition.X
		}
		if transition.Y > maxY {
			maxY = transition.Y
		}
	}
	return maxX + padding, maxY + padding
}

func (model *Model) ToSvg() string {
	var sb strings.Builder
	width, height := model.CanvasSize()

	sb.WriteString(ufmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%f" height="%d" viewBox="0 0 %d %d">\n`,
		width, height, width, height))

	sb.WriteString(`
        <style>
            .place { fill: #f0f0f0; stroke: #000000; stroke-width: 2; }
            .transition { fill: #ffffff; stroke: #000000; stroke-width: 2; }
            .arc { stroke: #000000; stroke-width: 1; }
            .label { font-family: Arial, sans-serif; font-size: 14px; fill: #000000; }
            .label-background { fill: #f0f0f0; stroke: #000000; stroke-width: 1; }
            .tokenSmaller { font-size: 12px; }
        </style>
    `)

	sb.WriteString(`
        <defs>
            <marker id="arrow" markerWidth="22.5" markerHeight="12" refX="9" refY="6.5" orient="auto">
                <path d="M3,1.5 L3,12 L10.5,6 L3,1.5"/>
            </marker>
            <marker id="inhibit" markerWidth="30" markerHeight="16" refX="10" refY="8.5" orient="auto">
                <circle cx="8" cy="9" r="4"/>
            </marker>
        </defs>
    `)

	for _, arrow := range model.Arrows {
		var sourceX, sourceY, targetX, targetY int

		// Determine source coordinates
		if place, ok := model.Places[arrow.Source]; ok {
			sourceX, sourceY = place.X, place.Y
		} else if transition, ok := model.Transitions[arrow.Source]; ok {
			sourceX, sourceY = transition.X, transition.Y
		} else {
			continue // Skip if source is invalid
		}

		// Determine target coordinates
		if place, ok := model.Places[arrow.Target]; ok {
			targetX, targetY = place.X, place.Y
		} else if transition, ok := model.Transitions[arrow.Target]; ok {
			targetX, targetY = transition.X, transition.Y
		} else {
			continue // Skip if target is invalid
		}

		// Calculate shortened coordinates
		dx := targetX - sourceX
		dy := targetY - sourceY
		length := math.Sqrt(float64(dx*dx + dy*dy))
		shortenFactor := 24
		x2 := targetX - int(float64(dx)/length*float64(shortenFactor))
		y2 := targetY - int(float64(dy)/length*float64(shortenFactor))

		marker := "arrow"
		if arrow.Inhibit {
			marker = "inhibit"
		}

		sb.WriteString(ufmt.Sprintf("<line x1=\"%d\" y1=\"%d\" x2=\"%d\" y2=\"%d\" class=\"arc\" marker-end=\"url(#%s)\" />\n",
			sourceX, sourceY, x2, y2, marker))

		// Add mid-marker circle and weight
		midX := (sourceX + x2) / 2
		midY := (sourceY + y2) / 2

		if len(arrow.Weight) == 0 || arrow.Weight[0] == 0 {
			continue // Skip if no weight
		}

		className := "label"
		if arrow.Weight[0] > 1000 {
			className = "tokenSmaller"
		}

		weightStr := fmtWeight(arrow.Weight)
		sb.WriteString(ufmt.Sprintf("<circle cx=\"%d\" cy=\"%d\" r=\"14\" class=\"label-background\" />\n", midX, midY))
		sb.WriteString(ufmt.Sprintf("<text x=\"%d\" y=\"%d\" class=\"%s\">%s</text>\n", midX-11, midY+5, className, weightStr))
	}

	// Draw places
	for _, place := range model.Places {
		sb.WriteString(ufmt.Sprintf("<circle cx=\"%d\" cy=\"%d\" r=\"16\" class=\"place\" />\n", place.X, place.Y))
		sb.WriteString(ufmt.Sprintf("<text x=\"%d\" y=\"%d\" class=\"label\">%s</text>\n", place.X-18, place.Y-20, place.Label))
	}

	// Draw transitions
	for i, transition := range model.Transitions {
		label := ufmt.Sprintf("%s", i)
		sb.WriteString(ufmt.Sprintf("<rect x=\"%d\" y=\"%d\" width=\"30\" height=\"30\" class=\"transition\" />\n", transition.X-15, transition.Y-15))
		sb.WriteString(ufmt.Sprintf("<text x=\"%d\" y=\"%d\" class=\"label\">%s</text>\n", transition.X-15, transition.Y-20, label))
	}

	sb.WriteString("</svg>")
	return sb.String()
}

func (model *Model) LegendDataUrl() string {
	return "data:image/svg+xml;utf8," + url.PathEscape(model.LegendSvg())
}

var legendStyles = `
<style>
    table { font-family: Arial, sans-serif; font-size: 16px; border-collapse: collapse; width: 100%;}
    th, td { border: 1px solid #ccc; padding: 6px 12px; text-align: left; }
    th { background: #f0f0f0; }
    tr:nth-child(even) { background: #fafafa; }
    td:nth-child(2) { max-width: 180px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
</style>
`

func (model *Model) LegendSvg() string {
	var sb strings.Builder
	rowHeight := 37
	numRows := len(model.Places) + len(model.Transitions) + len(model.Arrows) + len(model.Objects) + 2 // Add rows for objects, header, and hash
	height := numRows * rowHeight
	if numRows < 10 {
		height = 2*rowHeight + height
	}
	width := 700

	sb.WriteString(ufmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%v" height="%v">`, width, height))
	sb.WriteString(legendStyles)
	sb.WriteString(`<foreignObject x="10" y="10" width="680" height="` + ufmt.Sprintf("%d", height-20) + `">`)
	sb.WriteString(`<div xmlns="http://www.w3.org/1999/xhtml">`)
	sb.WriteString(`<table><thead><tr><th>Type</th><th>Label</th><th>Initial</th><th>Capacity</th><th>Weight</th></tr></thead><tbody>`)

	// Add objects section
	sb.WriteString(`<tr><td colspan="5"><strong>Objects:</strong></td></tr>`)
	for i, obj := range model.Objects {
		sb.WriteString(ufmt.Sprintf(`<tr><td colspan="5">%d: %s</td></tr>`, i, obj))
	}

	// Add places
	for label, place := range model.Places {
		sb.WriteString(`<tr><td>Place</td><td>` + label + `</td><td>` + place.Initial.String() + `</td><td>` + place.Capacity.String() + `</td><td>-</td></tr>`)
	}

	// Add transitions
	for label := range model.Transitions {
		sb.WriteString(`<tr><td>Transition</td><td>` + label + `</td><td>-</td><td>-</td><td>-</td></tr>`)
	}

	// Add arrows
	for _, arrow := range model.Arrows {
		arrowType := "Arc"
		if arrow.Inhibit {
			arrowType = "Inhibit"
		}
		sb.WriteString(`<tr><td>` + arrowType + `</td><td>` + arrow.Source + ` → ` + arrow.Target + `</td><td>-</td><td>-</td><td>` + fmtWeight(arrow.Weight) + `</td></tr>`)
	}

	// Add structural hash
	sb.WriteString(`<tr><td colspan="5"><strong>ID:</strong> ` + model.Cid() + `</td></tr>`)

	sb.WriteString(`</tbody></table></div></foreignObject>`)
	sb.WriteString("</svg>")
	return sb.String()
}

func (model *Model) DataUrl() string {
	return "data:image/svg+xml;utf8," + url.PathEscape(model.ToSvg())
}

func (model *Model) ThumbnailDataUrl() string {
	return "data:image/svg+xml;utf8," + url.PathEscape(model.ThumbnailSvg("140", "140"))
}

func (model *Model) ThumbnailSvg(w, h string) string {
	dataUrl := model.DataUrl()
	return `<svg width="` + w + `" height="` + h + `" xmlns="http://www.w3.org/2000/svg">
        <rect width="100%" height="100%" fill="none" stroke="#facade"/>
        <image href="` + dataUrl + `" width="` + w + `" height="` + h + `"/>
    </svg>`
}

func (model *Model) Markdown() string {
	var sb strings.Builder

	sb.WriteString("![model viz](")
	sb.WriteString(model.DataUrl())
	sb.WriteString(")\n\n")

	sb.WriteString("![model legend](")
	sb.WriteString(model.LegendDataUrl())
	sb.WriteString(")\n\n")

	return sb.String()
}

func AssertValidObjectName(name string) {
	if !strings.HasPrefix(name, "$") {
		panic("Invalid object name: " + name + ". Object names must start with '$'.")
	}
}

func (model *Model) Cid() string {
	return shaToCid(model.IdentityHash())
}

func (model *Model) IdentityHash() string {
	var elements []string

	sort.Strings(model.Objects)
	for _, obj := range model.Objects {
		elements = append(elements, hashString(obj))
	}

	placeKeys := make([]string, 0, len(model.Places))
	for label := range model.Places {
		placeKeys = append(placeKeys, label)
	}
	sort.Strings(placeKeys)
	for _, label := range placeKeys {
		elements = append(elements, hashString(label))
	}

	transitionKeys := make([]string, 0, len(model.Transitions))
	for label := range model.Transitions {
		transitionKeys = append(transitionKeys, label)
	}
	sort.Strings(transitionKeys)
	for _, label := range transitionKeys {
		elements = append(elements, hashString(label))
	}

	for _, arrow := range model.Arrows {
		if arrow.Inhibit {
			arcHash := hashString(arrow.Source + "-|>" + arrow.Target)
			elements = append(elements, arcHash)
		} else {
			arcHash := hashString(arrow.Source + "-->" + arrow.Target)
			elements = append(elements, arcHash)
		}
	}

	return buildMerkleRoot(elements)
}

// hashString generates a SHA-256 hash for a string.
func hashString(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// shaToCid converts a hex-encoded SHA-256 digest into a CIDv1 (base32, multibase "b").
// Layout: <0x01 version><0x55 codec=raw><0x12 mh=sha2-256><0x20 len=32><digest>
func shaToCid(sha string) string {
	sha = strings.TrimSpace(sha)

	// Decode hex digest (must be 32 bytes for sha2-256)
	digest, err := hex.DecodeString(sha)
	if err != nil || len(digest) != 32 {
		panic("invalid sha256 digest")
	}

	// Assemble CIDv1 bytes
	// 0x01 = version 1
	// 0x55 = multicodec "raw"
	// 0x12 = multihash code for sha2-256, 0x20 = digest length (32)
	buf := make([]byte, 0, 1+1+2+32)
	buf = append(buf, 0x01, 0x55, 0x12, 0x20)
	buf = append(buf, digest...)

	// Base32 (RFC 4648), no padding, lowercase, with multibase 'b' prefix
	enc := base32.StdEncoding.WithPadding(base32.NoPadding)
	return "b" + strings.ToLower(enc.EncodeToString(buf))
}

// buildMerkleRoot builds the Merkle root from a list of hashes.
func buildMerkleRoot(hashes []string) string {
	if len(hashes) == 0 {
		return ""
	}
	for len(hashes) > 1 {
		var newLevel []string
		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				combined := hashes[i] + hashes[i+1]
				newLevel = append(newLevel, hashString(combined))
			} else {
				newLevel = append(newLevel, hashes[i]) // Carry forward the last hash if odd
			}
		}
		hashes = newLevel
	}
	return hashes[0]
}



/*
Definition (informal):
A system is monotone if “having more resources cannot disable behavior you could already do.”

Monotonicity means: if a transition is enabled under marking M,
it remains enabled under any "larger" marking M' (with more tokens).
Ordinary Petri nets are monotone; inhibitor arcs are not.
Adding tokens can *disable* an inhibitor, so adjacency cannot
be safely composed through them — transitivity breaks.

*/

/*
Inhibitor arcs intentionally break transitivity.

Reason: normal arcs are monotone (more tokens ⇒ still enabled), so collapsing
through them preserves “can-reach” structure. Inhibitors are anti-monotone:
adding tokens can *disable* a transition.

If you allowed them in ForgetPlaces/ForgetTransitions, you would create edges that do not hold
behaviorally. To avoid spurious adjacency, inhibitors are skipped.

Baez, Courser, Master, Shulman — Compositionality of Petri Nets / Open Petri Nets (2018–2022).

They highlight that algebraic laws (like transitivity via composition) hold for standard nets with monotone semantics,
but need adjustment or can outright fail in extensions like inhibitors.
*/


/*
Example: inhibitors break monotonicity (Esparza & Nielsen 1994).

Place p has an inhibitor arc to transition t.
- At marking M = [p=0], t is enabled.
- At marking M' = [p=1], t is disabled.

So M ≤ M' (M' has more tokens), but enabled(M) ⊄ enabled(M').
This violates monotonicity.

Consequence:
- With ordinary arcs, Petri nets are monotone and many problems are decidable
  (reachability is EXPSPACE-complete).
- With inhibitors, monotonicity fails; nets can simulate counter machines.
- Result: reachability and related problems become *undecidable*.
  (Esparza & Nielsen, "Decidability Issues for Petri Nets with Inhibitor Arcs", 1994)

This is why ForgetPlaces/ForgetTransitions explicitly skip inhibitor arcs:
composing through them would create edges that don’t exist behaviorally.

Note: many practical "inhibit" conditions can be rephrased monotonically,
e.g. by encoding them as initial markings or one-shot control places.
In those cases the system stays monotone and Forget* projections are safe.
True dynamic inhibitors (tests for zero that can later be refilled)
remain non-monotone and break transitivity; we skip them.

*/


/*
ForgetTransitions projects a P/T net onto its places.

Semantics:
- Treat each transition as hidden/internal and connect every input place of that
  transition to every output place of that transition.
- The result is a place–adjacency graph: an edge p1 → p2 exists iff some
  transition t has p1 → t and t → p2 (ignoring inhibits).

Use when:
- You want “token-flow potential” between places (state-space skeleton),
  e.g., for coarse reachability heuristics, layout hints, or structural summaries.

Notes:
- Drops enabling/consumption/production details and rates/weights (structure-only).
- Preserves topological connectivity but not behavioral properties.
- Inhibit arcs are intentionally ignored so they don’t fabricate flow.

Related:
- Classical transition elimination and net-reduction rules (Berthelot/Murata).
- Compositional abstraction as “hiding” internal transitions (open nets).
*/
func (model *Model) ForgetTransitions() *Model {
	newModel := &Model{
		Places:      map[string]Place{},
		Transitions: map[string]Transition{},
		Arrows:      []Arrow{},
	}

	// Copy places
	for label, place := range model.Places {
		newModel.Places[label] = place
	}

	// Find all Place -> Transition and Transition -> Place arrows
	pt := map[string][]string{} // transition: sources (places)
	tp := map[string][]string{} // transition: targets (places)
	for _, arrow := range model.Arrows {
		if arrow.Inhibit {
			continue // Skip inhibit arcs which intentionally break transitivity
		}
		if _, srcIsPlace := model.Places[arrow.Source]; srcIsPlace {
			if _, tgtIsTransition := model.Transitions[arrow.Target]; tgtIsTransition {
				pt[arrow.Target] = append(pt[arrow.Target], arrow.Source)
			}
		}
		if _, srcIsTransition := model.Transitions[arrow.Source]; srcIsTransition {
			if _, tgtIsPlace := model.Places[arrow.Target]; tgtIsPlace {
				tp[arrow.Source] = append(tp[arrow.Source], arrow.Target)
			}
		}
	}

	// For each transition, link all input places to all output places
	for t, sources := range pt {
		targets := tp[t]
		for _, src := range sources {
			for _, tgt := range targets {
				newModel.Arrows = append(newModel.Arrows, Arrow{
					Source:  src,
					Target:  tgt,
					Inhibit: false,
				})
			}
		}
	}

	return newModel
}
/*
ForgetPlaces projects a P/T net onto its transitions.

Semantics:
- Treat each place as hidden/internal and connect every input transition of that
  place to every output transition of that place.
- The result is a transition–adjacency graph (TAR): an edge t1 → t2 exists iff
  there is some place p with t1 → p and p → t2 (ignoring inhibits).

Use when:
- You want the “can-follow” relation between transitions (control-flow skeleton),
  e.g., for partial-order analysis, concurrency visualization, or quick pathing.

Notes:
- Drops token semantics, capacities, and markings (structure-only abstraction).
- Preserves causal adjacency but not reachability/liveness/boundedness.
- Inhibit arcs are intentionally ignored so they don’t fabricate adjacency.

Related:
- Transition Adjacency Relations (TAR) in workflow/Petri-net analysis.
- Nets-with-boundaries/compositional “hiding” (open nets / cospans).
*/
func (model *Model) ForgetPlaces() *Model {
	newModel := &Model{
		Places:      map[string]Place{},
		Transitions: map[string]Transition{},
		Arrows:      []Arrow{},
	}

	// Copy transitions
	for label, transition := range model.Transitions {
		newModel.Transitions[label] = transition
	}

	// Find all Transition -> Place and Place -> Transition arrows
	tp := map[string][]string{} // place: input transitions
	pt := map[string][]string{} // place: output transitions
	for _, arrow := range model.Arrows {
		if arrow.Inhibit {
			continue // Skip inhibit arcs which intentionally break transitivity
		}
		if _, srcIsTransition := model.Transitions[arrow.Source]; srcIsTransition {
			if _, tgtIsPlace := model.Places[arrow.Target]; tgtIsPlace {
				tp[arrow.Target] = append(tp[arrow.Target], arrow.Source)
			}
		}
		if _, srcIsPlace := model.Places[arrow.Source]; srcIsPlace {
			if _, tgtIsTransition := model.Transitions[arrow.Target]; tgtIsTransition {
				pt[arrow.Source] = append(pt[arrow.Source], arrow.Target)
			}
		}
	}

	// For each place, link all input transitions to all output transitions
	for place := range model.Places {
		inputs := tp[place]
		outputs := pt[place]
		for _, src := range inputs {
			for _, tgt := range outputs {
				newModel.Arrows = append(newModel.Arrows, Arrow{
					Source:  src,
					Target:  tgt,
					Inhibit: false,
				})
			}
		}
	}

	return newModel
}

/* References
- Esparza, J., & Nielsen, M. (1994). Decidability Issues for Petri Nets (BRICS Report RS-94-8). University of Aarhus. [Inhibitor arcs → undecidability]

- Baez, J. C., & Master, J. (2018). Open Petri Nets. arXiv preprint.
- Baez, J. C., & Master, J. (2020). Open Petri Nets. Mathematical Structures in Computer Science, 30(3), 314–341. [Compositional semantics via double categories]
*/