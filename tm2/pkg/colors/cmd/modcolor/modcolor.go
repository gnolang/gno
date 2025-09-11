package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/gnolang/gno/tm2/pkg/colors"
)

var (
	order  = []colors.Color{colors.None, colors.Gray, colors.Cyan, colors.Blue, colors.Green, colors.Yellow, colors.Red, colors.Magenta}
	modPtr = flag.Int("mod", 2, "modulo number of lines; maximum 8")
)

func main() {
	flag.Parse()
	if *modPtr < 2 || 8 < *modPtr {
		panic("--mod must be between 2 and 8")
	}

	mod := *modPtr
	rin := bufio.NewReader(os.Stdin)
	for i := 0; ; i++ {
		line, err := rin.ReadString('\n')
		if err == io.EOF {
			return
		} else if err != nil {
			panic(err)
		}
		color := order[i%mod]
		fmt.Println(color(line[:len(line)-1]))
	}
}
