package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"reflect"

	"github.com/tendermint/classic/abci/types"
	cmn "github.com/tendermint/classic/libs/common"
	"github.com/tendermint/go-amino-x"
)

func main() {

	conn, err := cmn.Connect("unix://test.sock")
	if err != nil {
		log.Fatal(err.Error())
	}

	// Make a bunch of requests
	counter := 0
	for i := 0; ; i++ {
		req := abci.RequestEcho{Message: "foobar"}
		_, err := makeRequest(conn, req)
		if err != nil {
			log.Fatal(err.Error())
		}
		counter++
		if counter%1000 == 0 {
			fmt.Println(counter)
		}
	}
}

func makeRequest(conn net.Conn, req abci.Request) (abci.Response, error) {
	var bufWriter = bufio.NewWriter(conn)

	// Write desired request
	_, err := amino.MarshalAnySizedWriter(bufWriter, req)
	if err != nil {
		return nil, err
	}
	_, err = amino.MarshalAnySizedWriter(bufWriter, abci.RequestFlush{})
	if err != nil {
		return nil, err
	}
	err = bufWriter.Flush()
	if err != nil {
		return nil, err
	}

	// Read desired response
	var res abci.Response
	_, err = amino.UnmarshalSizedReader(conn, &res, 0)
	if err != nil {
		return nil, err
	}
	var resFlush abci.Response
	_, err = amino.UnmarshalSizedReader(conn, &res, 0)
	if err != nil {
		return nil, err
	}
	if _, ok := resFlush.(abci.ResponseFlush); !ok {
		return nil, fmt.Errorf("Expected flush response but got something else: %v", reflect.TypeOf(resFlush))
	}

	return res, nil
}
