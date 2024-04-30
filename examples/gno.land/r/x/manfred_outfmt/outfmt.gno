package outfmt

import (
	"encoding/json"
	"strings"

	"gno.land/p/demo/rand"
	"gno.land/p/demo/ufmt"
)

type Result struct {
	Number  int
	Text    string
	Numbers []int
}

func (res *Result) String() string {
	output := ufmt.Sprintf("Text: %s\n", res.Text)
	output += ufmt.Sprintf("Number: %d\n", res.Number)

	output += "Numbers:"
	for _, number := range res.Numbers {
		output += ufmt.Sprintf(" %d", number)
	}
	output += "\n"

	return output
}

var rSeed int64

func genResult() Result {
	// init rand
	r := rand.FromSeed(rSeed)
	defer func() { rSeed = r.Seed() }()

	res := Result{
		Text:   "Hello Gnomes!",
		Number: r.Intn(1000),
	}

	length := r.Intn(8) + 2
	res.Numbers = make([]int, length)
	for i := 0; i < length; i++ {
		res.Numbers[i] = r.Intn(100)
	}

	return res
}

func Render(path string) string {
	if path == "" {
		output := ""
		output += "* [?fmt=stringer](/r/x/manfred_outfmt:?fmt=stringer)\n"
		output += "* [?fmt=json](/r/x/manfred_outfmt:?fmt=json)\n"
		output += "* [?fmt=jsonp](/r/x/manfred_outfmt:?fmt=jsonp)\n"
		return output
	}

	// parsing
	fmt := "stringer" // default
	// FIXME: whitelist and use net/url.
	{
		parts := strings.Split(path, "?")
		if len(parts) == 2 {
			// path := parts[0]
			query := parts[1]
			if query != "" {
				parts = strings.Split(query, "=")
				if len(parts) == 2 {
					k := parts[0]
					v := parts[1]
					if k == "fmt" {
						fmt = v
					}
				}
			}
		}
	}

	// rendering

	res := genResult()
	switch fmt {
	case "stringer":
		return res.String()
	case "json":
		// FIXME: return content-type?
		out, err := json.Marshal(res)
		checkErr(err)
		return string(out)
	case "jsonp":
		// FIXME: return content-type?
		callback := "callback" // FIXME: get it from query
		out, err := json.Marshal(res)
		checkErr(err)
		return callback + "(" + string(out) + ")"
	default:
		panic("unknown format")
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
