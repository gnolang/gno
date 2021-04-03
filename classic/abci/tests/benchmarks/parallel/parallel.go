package main

import (
	"bufio"
	"fmt"
	"log"

	"github.com/tendermint/classic/abci/types"
	cmn "github.com/tendermint/classic/libs/common"
	"github.com/tendermint/go-amino-x"
)

func main() {

	const maxSize = 1e6

	conn, err := cmn.Connect("unix://test.sock")
	if err != nil {
		log.Fatal(err.Error())
	}

	// Read a bunch of responses
	go func() {
		counter := 0
		for {
			var res abci.Response
			_, err := amino.UnmarshalSizedReader(conn, &res, maxSize)
			if err != nil {
				log.Fatal(err.Error())
			}
			counter++
			if counter%1000 == 0 {
				fmt.Println("Read", counter)
			}
		}
	}()

	// Write a bunch of requests
	counter := 0
	for i := 0; ; i++ {
		var bufWriter = bufio.NewWriter(conn)
		var req = abci.RequestEcho{Message: "foobar"}

		_, err := amino.MarshalAnySizedWriter(bufWriter, req)
		err = bufWriter.Flush()
		if err != nil {
			log.Fatal(err.Error())
		}

		counter++
		if counter%1000 == 0 {
			fmt.Println("Write", counter)
		}
	}
}
