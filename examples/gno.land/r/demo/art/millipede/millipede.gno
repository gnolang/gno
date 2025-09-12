package millipede

import (
	"strconv"
	"strings"

	"gno.land/p/nt/ufmt"
)

const (
	minSize     = 1
	defaultSize = 20
	maxSize     = 100
)

func Draw(size int) string {
	if size < minSize || size > maxSize {
		panic("invalid millipede size")
	}
	paddings := []string{"  ", " ", "", " ", "  ", "   ", "    ", "    ", "   "}
	var b strings.Builder
	b.WriteString("    ╚⊙ ⊙╝\n")
	for i := 0; i < size; i++ {
		b.WriteString(paddings[i%9] + "╚═(███)═╝\n")
	}
	return b.String()
}

func Render(path string) string {
	size := defaultSize

	path = strings.TrimSpace(path)
	if path != "" {
		var err error
		size, err = strconv.Atoi(path)
		if err != nil {
			panic(err)
		}
	}

	output := "```\n" + Draw(size) + "```\n"
	if size > minSize {
		output += ufmt.Sprintf("[%d](/r/demo/art/millipede:%d)< ", size-1, size-1)
	}
	if size < maxSize {
		output += ufmt.Sprintf(" >[%d](/r/demo/art/millipede:%d)", size+1, size+1)
	}
	return output
}

// based on https://github.com/getmillipede/millipede-go/blob/977f046c39c35a650eac0fd30245e96b22c7803c/main.go
