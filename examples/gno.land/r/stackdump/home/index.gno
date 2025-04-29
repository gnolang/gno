package home

import (
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

var (
	registry    = avl.NewTree()
	render []string
)

func SetOutput(s ...string) {
    render = []string{}
    outputOrder(s...)
}

func outputOrder(s ...string) {
    for _, v := range s {
        render = append(render, v)
    }
}

func init() {
	outputOrder("foo", "bar")
    registry.Set("foo", "testing!")
    registry.Set("bar", "\n123")
}

// REVIEW: used in init() without cross()
func Set(key string, val interface{}) {
	  //crossing() 
    registry.Set(key, val)
}

func SetString(key string, val string) {
    registry.Set(key, val)
}

func GetString(key string) string {
    val, _ := registry.Get(key)
    switch val.(type) {
    case string, int, float64, bool:
        return ufmt.Sprintf("%v", val)
    default:
        return ""
    }
}

func Render(_ string) string {
    out := strings.Builder{}
    for _, key := range render {
        val := GetString(key)
        if val != "" {
            out.WriteString(val)
        }
    }
    return out.String()
}

