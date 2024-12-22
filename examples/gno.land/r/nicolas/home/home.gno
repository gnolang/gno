package home

import (
	"strings"
    "gno.land/p/demo/ufmt"
    
	"gno.land/r/nicolas/config"
    "gno.land/r/leon/hof"
)

func init() {
    hof.Register()
}


const HTML = "# Nicolas's (gn)home Dashboard"

func Render(path string) string {
	return HTML
}