package calculator

import (
	"net/url"
	"strconv"
	"strings"

	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/realmpath"
	"gno.land/r/leon/hor"
)

type Node struct {
	value string // Value of the current node
	left  *Node
	right *Node
}

const (
	specialCharacters             = "+-*/."
	specialCharactersWithoutMinus = "+*/."
	topPriority                   = "*/"
	lowPriority                   = "+-"
	realmPath                     = "/r/miko/calculator"
)

var (
	val        float64
	displayVal string

	operationMap = map[string]func(left float64, right float64) float64{
		"+": func(left float64, right float64) float64 { return left + right },
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
	hor.Register(cross, "Miko's calculator", "")
}

func evaluateValidity(line string) (bool, string) {
	if len(line) == 0 {
		return false, "Invalid empty input"
	} // edge case empty line
	if strings.Index(specialCharactersWithoutMinus, string(line[0])) != -1 ||
		strings.Index(specialCharacters, string(line[len(line)-1])) != -1 {
		return false, "Invalid equation"
	} // edge case special character at begining or end

	isPriorSpecial := false
	countParenthesis := 0

	for i := 0; i < len(line); i++ {
		if line[i] == '(' {
			countParenthesis += 1
			continue
		}
		if line[i] == ')' {
			if isPriorSpecial == true {
				return false, "Invalid equation"
			}
			countParenthesis -= 1
			isPriorSpecial = false
			continue
		}
		if strings.Index(specialCharacters, string(line[i])) != -1 {
			if isPriorSpecial && !(line[i] == '-' && i < (len(line)-1) && line[i+1] >= '0' && line[i+1] <= '9') { // if we have two subsequent operator and the second one isn't a - before a number (negative number)
				return false, "Invalid equation"
			}
			isPriorSpecial = true
			continue
		}
		if line[i] < '0' || line[i] > '9' {
			return false, "Invalid character encountered "
		}
		isPriorSpecial = false
	}

	if countParenthesis != 0 {
		return false, "Invalid equation"
	}
	return true, ""
}

func searchForPriority(priorityList string, line string) *Node {
	countParenthesis := 0
	for iPrio := 0; iPrio < len(priorityList); iPrio++ {
		for idx := 0; idx < len(line); idx++ {
			if line[idx] == '(' {
				countParenthesis += 1
			}
			if line[idx] == ')' {
				countParenthesis -= 1
			}
			if countParenthesis == 0 && line[idx] == priorityList[iPrio] &&
				!(line[idx] == '-' && (idx == 0 || strings.Index(specialCharacters, string(line[idx-1])) != -1)) { // - is not a substract sign if at the begining or after another sign
				return &Node{string(line[idx]), createTree(line[:idx]), createTree(line[idx+1:])}
			}

		}
	}
	return nil
}

// checks if the expression in line is contained in one big parenthesis
func isInOneParenthesis(line string) bool {
	if line[0] != '(' || line[len(line)-1] != ')' {
		return false
	}
	countParenthesis := 1
	for i := 1; i < len(line)-1; i++ {
		if line[i] == '(' {
			countParenthesis += 1
		}
		if line[i] == ')' {
			countParenthesis -= 1
		}
		if countParenthesis == 0 {
			return false
		}
	}
	return true
}

func createTree(line string) *Node {
	if isInOneParenthesis(line) {
		return createTree(line[1 : len(line)-1])
	}
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

func createTable(req *realmpath.Request, query url.Values, expression string) mdtable.Table {
	line := make([]string, 0, 4)
	query.Set("expression", "")
	line = append(line, md.Link("res", req.String()))
	for _, str := range []string{"(", ")"} {
		query.Set("expression", expression+str)
		line = append(line, md.Link(str, req.String()))
	}
	query.Set("expression", removeLast(expression))
	line = append(line, md.Link("del", req.String())) // req and del are two special cases
	table := mdtable.Table{
		Headers: line,
	}
	line = []string{}
	for _, c := range "789+456-123*0.=/" {
		query.Set("expression", expression+string(c))
		line = append(line, md.Link(string(c), req.String()))
		if len(line) == 4 {
			table.Append(line)
			line = []string{}
		}
	}
	return table
}

func Render(path string) string {
	req := realmpath.Parse(path)
	query := req.Query
	expression := query.Get("expression")

	if expression == "" {
		displayVal = "0"
	} else {
		if expression[len(expression)-1] == '=' {
			expression = ComputeResult(expression[:len(expression)-1])
		} else {
			displayVal = expression
		}
	}
	out := md.H1("Calculator page")
	out += md.H3("Have you ever wanted to do maths but never actually found a calculator ?")
	out += md.H3("Do I have the realm for you...")
	out += "---------------\n"
	out += md.H2("Result: " + displayVal)

	table := createTable(req, query, expression)
	out += table.String()
	return out
}
