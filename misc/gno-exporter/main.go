package main

import (
	"flag"
	"fmt"

	gnocollectors "github.com/gnolang/gno/misc/gno-exporter/gno-collectors"
)

var (
	rpcURL = flag.String("node", "", "gno node url")
)

func main() {
	flag.Parse()

	if *rpcURL == "" {
		fmt.Println("usage: gno-exporter -node <node_url>")
		return
	}

	c := gnocollectors.NewGnoCollector(gnocollectors.GnoCollectorOpts{
		RPCURL:     *rpcURL,
		Collectors: []gnocollectors.Collector{},
	})
	c.AddCollectors()

	if err := c.Start(":8080"); err != nil {
		fmt.Println(err)
	}
}
