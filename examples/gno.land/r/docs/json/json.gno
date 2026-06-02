// Package json shows how to build and read JSON in Gno.
// It uses the /p/onbloc/json package.
package json

import (
	"gno.land/p/onbloc/json"
)

// lastCreatedJSON stores the most recently created JSON for the live demo.
var lastCreatedJSON string

// CreateUserJSON builds a JSON object from user data.
func CreateUserJSON(_ realm, name, email string, age int) string {
	return encode(json.ObjectNode("", map[string]*json.Node{
		"name":  json.StringNode("name", name),
		"email": json.StringNode("email", email),
		"age":   json.NumberNode("age", float64(age)),
	}))
}

// CreateArrayJSON builds a JSON array from the given items.
func CreateArrayJSON(_ realm, items ...string) string {
	nodes := make([]*json.Node, len(items))
	for i, item := range items {
		nodes[i] = json.StringNode("", item)
	}
	return encode(json.ArrayNode("", nodes))
}

// encode marshals the node, records it for the demo, and returns the JSON.
func encode(node *json.Node) string {
	out, err := json.Marshal(node)
	if err != nil {
		panic(err)
	}
	lastCreatedJSON = string(out) // global, shown by the live demo
	return lastCreatedJSON
}

func Render(_ string) string {
	return `# JSON in Gno

Done through the [` + "`/p/onbloc/json`" + `](/p/onbloc/json) package.

## Creating JSON Objects

` + "```go" + `
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

` + "```go" + `
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

` + "```go" + `
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
- ` + "`json.Object`" + ` - JSON objects ` + "`{}`" + `
- ` + "`json.Array`" + ` - JSON arrays ` + "`[]`" + `
- ` + "`json.String`" + ` - Strings
- ` + "`json.Number`" + ` - Numbers (integers and floats)
- ` + "`json.Boolean`" + ` - Booleans (true/false)
- ` + "`json.Null`" + ` - Null values

## Live Demo

` + renderDemo() + `

---

### See also

- [/p/onbloc/json](/p/onbloc/json) - JSON package documentation
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
	out += "[Create Sample Array JSON](/r/docs/json$help&func=CreateArrayJSON)\n"
	return out
}
