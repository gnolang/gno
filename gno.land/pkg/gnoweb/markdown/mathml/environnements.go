package mathml

import (
	"strconv"
	"strings"
)

// Isolate the environment context from the current context
func isolateEnvironmentContext(ctx parseContext) parseContext {
	return ctx & ((ctxVarNormal - 1) ^ (ctxTable - 1))
}

// Set the context to the environment context and return the new context
func setEnvironmentContext(envBegin Token, context parseContext) parseContext {
	context = context ^ isolateEnvironmentContext(context) // clear other environments
	star := strings.HasSuffix(envBegin.Value, "*")
	name := strings.TrimSuffix(envBegin.Value, "*")
	switch name {
	case "matrix", "pmatrix", "bmatrix", "Bmatrix", "vmatrix", "Vmatrix":
		if star {
			context |= ctxEnvHasArg
		}
		return context | ctxTable
	case "array", "subarray":
		return context | ctxTable | ctxEnvHasArg
	case "table", "align", "aligned", "cases":
		return context | ctxTable
	}
	return context
}

// split a slice whenever an element e of s satisfies f(e) == true.
func splitByFunc[T any](s []T, f func(T) bool) [][]T {
	out := make([][]T, 0)
	temp := make([]T, 0)
	if s != nil {
		for _, t := range s {
			if f(t) {
				out = append(out, temp)
				temp = make([]T, 0)
				continue
			}
			temp = append(temp, t)
		}
		if len(temp) > 0 {
			out = append(out, temp)
		}
	}
	return out
}

// remove duplicates from the end of a list
func trim(lst []string) []string {
	stop := len(lst)
	if len(lst) > 1 {
		val := lst[len(lst)-1]
		for stop = len(lst); stop > 0; stop-- {
			if lst[stop-1] != val {
				stop++
				break
			}
		}
		if stop <= 0 {
			stop = len(lst)
		}
	}
	return lst[:stop]
}

// take a string like "l|c|r" and produce the strings "left center right" and "solid solid",
// these being the values of the columnalign and colunlines properties respectively
// Note that mathml does not directly support drawing a line before the first or after the last column.
func parseAlignmentString(str string) ([]string, []string) {
	align := make([]string, 0, len(str))
	lines := make([]string, 0, len(str))
	wasline := true
	for i, c := range str {
		switch c {
		case 'l':
			align = append(align, "left")
		case 'c':
			align = append(align, "center")
		case 'r':
			align = append(align, "right")
		case '|':
			if i > 0 {
				lines = append(lines, "solid")
				wasline = true
			}
		case ':':
			if i > 0 {
				lines = append(lines, "dashed")
				wasline = true
			}
		}
		switch c {
		case 'l', 'c', 'r':
			if !wasline {
				lines = append(lines, "none")
			}
			wasline = false
		}
	}
	return trim(align), trim(lines)
}

// Process the table of environment env.
func processTable(table *MMLNode, env string) {
	if table == nil {
		return
	}
	table.SetAttr("columnalign", "center") // default
	align, lines := parseAlignmentString(table.Option)
	if len(align) > 0 {
		table.SetAttr("columnalign", strings.Join(align, " "))
	}
	if len(lines) > 0 {
		table.SetAttr("columnlines", strings.Join(lines, " "))
	}
	rows := make([]*MMLNode, 0)
	var cellNode *MMLNode
	rowspans := make(map[int]int)
	rowspacing := make([]string, 0)
	nonDefaultSpacing := false
	separateRows := func(n *MMLNode) bool { return n != nil && n.Properties&propRowSep > 0 }
	separateCells := func(n *MMLNode) bool { return n != nil && n.Properties&propCellSep > 0 }
	for _, row := range splitByFunc(table.Children, separateRows) {
		rowNode := NewMMLNode("mtr")
		space := "1.0ex"
		// cidx is the logical column of the next cell: a \multicolumn cell
		// is one cell of the source but covers several columns, and the
		// cell after it starts after the last of them.
		cidx := 0
		for _, cell := range splitByFunc(row, separateCells) {
			// A \multirow cell above covers this column, and the source
			// leaves its place empty: emit no <mtd> for it. If the place is
			// not empty, the cell is emitted anyway (and lands after the
			// spanning cell) rather than dropped.
			if rowspans[cidx] > 0 {
				rowspans[cidx]--
				if isEmptyCell(cell) {
					cidx++
					continue
				}
			}
			cellNode = NewMMLNode("mtd")
			cellNode.Children = append(cellNode.Children, cell...)
			if a := columnAlign(env, align, cidx); a != "" {
				cellNode.SetAttr("columnalign", a)
			}
			width := 1
			for i, c := range cell {
				if c == nil {
					continue
				}
				if s, ok := c.Attrib["rowspacing"]; ok {
					space = s
					nonDefaultSpacing = true
				}
				if spanstr, ok := c.Attrib["rowspan"]; ok {
					delete(cellNode.Children[i].Attrib, "rowspan")
					cellNode.SetAttr("rowspan", spanstr)
					span, err := strconv.ParseInt(spanstr, 10, 16)
					if err == nil {
						rowspans[cidx] = int(span) - 1
					}
					if len(cell) == 1 && c.Properties&propVertArrow > 0 {
						// rows have a default height of 1em and space of 1ex=½em between them.
						// There is one less interior space than the number of rows spanned.
						// total height of this combined cell:
						// span + (span-1)/2 = ((3*span)-1)/2
						minsize := float32((3*span)-1) / 2
						cellNode.Children[0].SetAttr("minsize", strconv.FormatFloat(float64(minsize), 'f', 1, 32)+"em")
					}
				}
				if spanstr, ok := c.Attrib["columnspan"]; ok {
					delete(cellNode.Children[i].Attrib, "columnspan")
					cellNode.SetAttr("columnspan", spanstr)
					span, err := strconv.ParseInt(spanstr, 10, 16)
					if err == nil {
						width = int(span)
					}
					if len(cell) == 1 && c.Properties&propHorzArrow > 0 {
						arrowWidth := strconv.FormatFloat(float64(2*span-1), 'f', 1, 32) + "em"
						// This is a "hack" for browsers compatibility. Arrows do not like to stretch.
						// Browsers should get this fixed soon.
						mover := NewMMLNode("mover")
						mspace := NewMMLNode("mspace")
						mspace.SetAttr("width", arrowWidth)
						mover.AppendChild(c, mspace)
						cellNode.Children[0] = mover
					}
				}
			}
			rowNode.AppendChild(cellNode)
			cidx += width
		}
		if nonDefaultSpacing {
			rowspacing = append(rowspacing, space)
		} else {
			rowspacing = append(rowspacing, "1.0ex")
		}
		rows = append(rows, rowNode)
	}
	if nonDefaultSpacing {
		table.SetAttr("rowspacing", strings.Join(trim(rowspacing), " "))
	}
	table.Tag = "mtable"
	table.SetAttr("rowalign", "center")
	table.Children = rows
}

// isEmptyCell reports whether a table cell holds nothing.
func isEmptyCell(cell []*MMLNode) bool {
	for _, n := range cell {
		if n != nil {
			return false
		}
	}
	return true
}

// columnAlign returns the alignment of column col of environment env, whose
// column spec parsed to align, or "" to leave the cell to the <mtd> default.
// Cells carry it as a columnalign attribute, which the stylesheet maps to
// text-align and padding. Empty cells keep it: their padding is what
// separates the column pairs of an aligned environment.
func columnAlign(env string, align []string, col int) string {
	env = strings.TrimSuffix(env, "*")
	switch env {
	case "cases":
		return "left"
	case "align", "aligned":
		// Columns pair up as right-aligned left side, left-aligned right side.
		return [...]string{"right", "left"}[col%2]
	}
	if len(align) > 0 {
		// The last alignment of the spec repeats for the remaining columns.
		return align[min(col, len(align)-1)]
	}
	switch env {
	case "pmatrix", "bmatrix", "Bmatrix", "vmatrix", "Vmatrix", "subarray":
		// Explicitly centered: Chrome's <mtd> default (-webkit-center)
		// lays out cells of uneven width differently.
		return "center"
	}
	return ""
}

// Create a strechy opening or closing parenthesis
func strechyOP(c string) *MMLNode {
	n := NewMMLNode("mo", c)
	n.SetAttr("stretchy", "true")
	n.SetAttr("fence", "true")
	return n
}

func processEnv(node *MMLNode, env string, ctx parseContext) *MMLNode {
	if ctx&ctxTable > 0 {
		processTable(node, env)
	}
	row := NewMMLNode("mrow")
	var left, right *MMLNode
	attrib := make(map[string]string)
	switch env {
	case "pmatrix", "pmatrix*":
		left = strechyOP("(")
		right = strechyOP(")")
	case "bmatrix", "bmatrix*":
		left = strechyOP("[")
		right = strechyOP("]")
	case "Bmatrix", "Bmatrix*":
		left = strechyOP("{")
		right = strechyOP("}")
	case "vmatrix", "vmatrix*":
		left = strechyOP("|")
		right = strechyOP("|")
	case "Vmatrix", "Vmatrix*":
		left = strechyOP("‖")
		right = strechyOP("‖")
	case "cases":
		left = strechyOP("{")
		right = NewMMLNode("mo") // Empty mo for proper fence pairing
	case "align", "align*", "aligned":
		attrib["displaystyle"] = "true"
		// The stylesheet pairs up the columns of an aligned table.
		attrib["class"] = "math-aligned"
	case "subarray":
		attrib["displaystyle"] = "false"
	default:
		return node
	}
	if node != nil {
		for k, v := range attrib {
			node.SetAttr(k, v)
		}
	}
	row.Children = append(row.Children, left, node, right)
	return row
}
