package calculator

import (
	"strconv"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/realmpath"
	"gno.land/r/leon/hof"
)

type Node struct {
	value string // Value of the current node
	left  *Node
	right *Node
}

const (
	specialCharacters = "p-*/."
	topPriority       = "*/"
	lowPriority       = "p-"

	realmPath = "/r/miko/calculator"
)

var (
	val        float64
	displayVal string

	operationMap = map[string]func(left float64, right float64) float64{
		"p": func(left float64, right float64) float64 { return left + right },
		"-": func(left float64, right float64) float64 { return left - right },
		"*": func(left float64, right float64) float64 { return left * right },
		"/": func(left float64, right float64) float64 {
			if right == 0 {
				panic("Division by 0 is forbidden")
			}
			return left / right
		},
	}
)

func init() {
	hof.Register("Miko's calculator", "Let's do maths")
}

func evaluateValidity(line string) (bool, string) {
	if len(line) == 0 {
		return false, "Invalid empty input"
	} // edge case empty line
	if strings.Index(specialCharacters, string(line[0])) != -1 ||
		strings.Index(specialCharacters, string(line[len(line)-1])) != -1 {
		return false, "Invalid equation"
	} // edge case special character at begining or end

	isPriorSpecial := false

	for i := 1; i < len(line)-1; i++ {
		if strings.Index(specialCharacters, string(line[i])) != -1 {
			if isPriorSpecial {
				return false, "Invalid equation"
			}
			isPriorSpecial = true
			continue
		}
		if (line[i] < '0' || line[i] > '9') && line[i] != 'p' {
			return false, "Invalid character encountered "
		}
		isPriorSpecial = false
	}
	return true, ""
}

func searchForPriority(priorityList string, line string) *Node {
	for i := 0; i < len(priorityList); i++ {
		idx := strings.Index(line, string(priorityList[i]))
		if idx != -1 {
			return &Node{string(line[idx]), createTree(line[:idx]), createTree(line[idx+1:])}
		}
	}
	return nil
}

func createTree(line string) *Node {
	node := searchForPriority(lowPriority, line) // we put the lowest priority at the top of the tree, these operations will be executed last
	if node != nil {
		return node
	}
	node = searchForPriority(topPriority, line)
	if node != nil {
		return node
	}

	// if this code is reached, the only value possible in line is a number
	return &Node{line, nil, nil}
}

func readTree(tree *Node) float64 {
	operation, exists := operationMap[tree.value]

	if exists { // check if the current node is an operator
		return operation(readTree(tree.left), readTree(tree.right))
	}

	parsedValue, _ := strconv.ParseFloat(tree.value, 64)

	return parsedValue
}

// expression is the equation you want to solve (p replaces the + symbol)
// exemple: 2p4/2
func ComputeResult(expression string) string {
	valid, errString := evaluateValidity(expression)

	if !valid { // If a basic error is encountered, return the expression without the = at the end and display the same expression
		println(errString) // display error for debug
		displayVal = strings.Replace(expression, "p", "+", -1)
		return expression
	}

	tree := createTree(expression)

	val = readTree(tree)
	displayVal = strconv.FormatFloat(val, 'g', 6, 64)
	return displayVal
}

func removeLast(path string) string {
	lenPath := len(path)
	if lenPath > 0 {
		path = path[:lenPath-1]
	}
	return path
}

func Render(path string) string {
	var sb strings.Builder

	req := realmpath.Parse(path)
	query := req.Query
	expression := query.Get("expression")

	if expression == "" {
		displayVal = "0"
	} else {
		if expression[len(expression)-1] == '=' {
			expression = ComputeResult(expression[:len(expression)-1])
		} else {
			displayVal = strings.Replace(expression, "p", "+", -1)
		}
	}

	sb.WriteString(`# Calculator page

Have you ever wanted to do maths but never actually found a calculator ?
Do I have the realm for you...

Result: ` + displayVal + `
---------------
| ` + ufmt.Sprintf("[res](%s)", realmPath) + `|||` + ufmt.Sprintf("[del](%s)", realmPath+":?expression="+removeLast(expression)) + `|
|---|---|---|---|
| ` + ufmt.Sprintf("[7](%s)", realmPath+":?expression="+expression+"7") + `| ` + ufmt.Sprintf("[8](%s)", realmPath+":?expression="+expression+"8") + `| ` + ufmt.Sprintf("[9](%s)", realmPath+":?expression="+expression+"9") + `| ` + ufmt.Sprintf("[+](%s)", realmPath+":?expression="+expression+"p") /* here p replaces + because of how + works in bnormal paths*/ + `|
| ` + ufmt.Sprintf("[4](%s)", realmPath+":?expression="+expression+"4") + `| ` + ufmt.Sprintf("[5](%s)", realmPath+":?expression="+expression+"5") + `| ` + ufmt.Sprintf("[6](%s)", realmPath+":?expression="+expression+"6") + `| ` + ufmt.Sprintf("[-](%s)", realmPath+":?expression="+expression+"-") + `|
| ` + ufmt.Sprintf("[1](%s)", realmPath+":?expression="+expression+"1") + `| ` + ufmt.Sprintf("[2](%s)", realmPath+":?expression="+expression+"2") + `| ` + ufmt.Sprintf("[3](%s)", realmPath+":?expression="+expression+"3") + `| ` + ufmt.Sprintf("[*](%s)", realmPath+":?expression="+expression+"*") + `|
| ` + ufmt.Sprintf("[0](%s)", realmPath+":?expression="+expression+"0") + `| ` + ufmt.Sprintf("[.](%s)", realmPath+":?expression="+expression+".") + `| ` + ufmt.Sprintf("[=](%s)", realmPath+":?expression="+expression+"=") + `| ` + ufmt.Sprintf("[/](%s)", realmPath+":?expression="+expression+"/") + `|
`)
	return sb.String()
}
