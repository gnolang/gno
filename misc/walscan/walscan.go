package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/wal"

	// Amino, amino, why so peculiar?
	_ "github.com/gnolang/gno/tm2/pkg/bft/consensus"
)

const (
	maxWALSize = 1 << 20 // 1MB
)

func main() {
	flag.Parse()

	filename := flag.Arg(0)
	if filename == "" {
		log.Fatalf("usage: %s <FILE>\n", flag.CommandLine.Name())
	}
	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("error opening %q: %v", filename, err)
	}
	defer f.Close()

	amino.RegisterPackage(amino.NewPackage("github.com/gnolang/gno/tm2/pkg/bft/wal", "wal", amino.GetCallersDirname()).WithTypes(
		&wal.MetaMessage{}, "MetaMessage",
		&wal.TimedWALMessage{}, "TimedMessage",
	))

	rd := wal.NewWALReader(f, maxWALSize)
	for {
		msg, meta, err := rd.ReadMessage()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			log.Fatalf("reading wal message: %v", err)
		}
		switch {
		case msg != nil:
			x := any(msg)
			fmt.Printf("%s\n", amino.MustMarshalJSON(&x))
		case meta != nil:
			x := any(meta)
			fmt.Printf("%s\n", amino.MustMarshalJSON(&x))
		default:
			panic("msg == meta == nil")
		}
	}
}
