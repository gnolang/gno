package json

import (
	"gno.land/p/onbloc/json"
)

// User represents a simple user structure
type User struct {
	Name  string
	Email string
	Age   int
}

// lastCreatedJSON stores the most recently created JSON for demo
var lastCreatedJSON string

// CreateUserJSON creates a JSON object from user data
func CreateUserJSON(_ realm, name, email string, age int) string {
	// Build JSON using the node builder pattern
	node := json.ObjectNode("", map[string]*json.Node{
		"name":  json.StringNode("name", name),
		"email": json.StringNode("email", email),
		"age":   json.NumberNode("age", float64(age)),
	})

	encoded, err := json.Marshal(node)
	if err != nil {
		panic("failed to marshal JSON: " + err.Error())
	}

	lastCreatedJSON = string(encoded)
	return lastCreatedJSON
}

// CreateArrayJSON creates a JSON array
func CreateArrayJSON(_ realm, items ...string) string {
	nodes := make([]*json.Node, len(items))
	for i, item := range items {
		nodes[i] = json.StringNode("", item)
	}

	node := json.ArrayNode("", nodes)
	encoded, err := json.Marshal(node)
	if err != nil {
		panic("failed to marshal JSON: " + err.Error())
	}

	lastCreatedJSON = string(encoded)
	return lastCreatedJSON
}

// ParseJSON parses a JSON string and returns formatted output
func ParseJSON(data string) (string, error) {
	node, err := json.Unmarshal([]byte(data))
	if err != nil {
		return "", err
	}

	// Pretty print the parsed JSON
	return prettyPrint(node, 0), nil
}

func prettyPrint(node *json.Node, indent int) string {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	switch node.Type() {
	case json.Object:
		result := "{\n"
		obj := node.MustObject()
		first := true
		for key, val := range obj {
			if !first {
				result += ",\n"
			}
			first = false
			result += prefix + "  \"" + key + "\": " + prettyPrint(val, indent+1)
		}
		result += "\n" + prefix + "}"
		return result
	case json.Array:
		result := "[\n"
		arr := node.MustArray()
		for i, val := range arr {
			if i > 0 {
				result += ",\n"
			}
			result += prefix + "  " + prettyPrint(val, indent+1)
		}
		result += "\n" + prefix + "]"
		return result
	case json.String:
		return "\"" + node.MustString() + "\""
	case json.Number:
		return node.String()
	case json.Boolean:
		if node.MustBool() {
			return "true"
		}
		return "false"
	case json.Null:
		return "null"
	default:
		return node.String()
	}
}

func Render(_ string) string {
	return `# JSON in Gno

JSON handling in Gno is done through the ` + "`/p/onbloc/json`" + ` package.

## Creating JSON Objects

Use the node builder pattern to create JSON:

` + "```gno" + `
import "gno.land/p/onbloc/json"

// Create a JSON object
node := json.ObjectNode("", map[string]*json.Node{
    "name":  json.StringNode("name", "Alice"),
    "email": json.StringNode("email", "alice@example.com"),
    "age":   json.NumberNode("age", 30),
})

// Serialize to string
encoded, err := json.Marshal(node)
// Result: {"name":"Alice","email":"alice@example.com","age":30}
` + "```" + `

## Creating JSON Arrays

` + "```gno" + `
items := []*json.Node{
    json.StringNode("", "apple"),
    json.StringNode("", "banana"),
    json.StringNode("", "cherry"),
}
node := json.ArrayNode("", items)

encoded, err := json.Marshal(node)
// Result: ["apple","banana","cherry"]
` + "```" + `

## Parsing JSON

` + "```gno" + `
data := ` + "`" + `{"name":"Bob","active":true}` + "`" + `
node, err := json.Unmarshal([]byte(data))
if err != nil {
    panic(err)
}

// Access values
name := node.MustKey("name").MustString()  // "Bob"
active := node.MustKey("active").MustBool() // true
` + "```" + `

## Node Types

The package supports these JSON types:
- ` + "`json.ObjectType`" + ` - JSON objects ` + "`{}`" + `
- ` + "`json.ArrayType`" + ` - JSON arrays ` + "`[]`" + `
- ` + "`json.StringType`" + ` - Strings
- ` + "`json.NumberType`" + ` - Numbers (integers and floats)
- ` + "`json.BoolType`" + ` - Booleans (true/false)
- ` + "`json.NullType`" + ` - Null values

## Nested Structures

` + "```gno" + `
address := json.ObjectNode("", map[string]*json.Node{
    "city":    json.StringNode("city", "New York"),
    "country": json.StringNode("country", "USA"),
})

user := json.ObjectNode("", map[string]*json.Node{
    "name":    json.StringNode("name", "Charlie"),
    "address": address,
})
` + "```" + `

## Live Demo

` + renderDemo() + `

---

See also: [p/onbloc/json](/p/onbloc/json)
`
}

func renderDemo() string {
	out := "### Last Created JSON\n\n"
	if lastCreatedJSON == "" {
		out += "_No JSON created yet. Try creating one!_\n\n"
	} else {
		out += "```json\n" + lastCreatedJSON + "\n```\n\n"
	}
	out += "[Create Sample User JSON](/r/docs/json$help&func=CreateUserJSON)\n"
	return out
}
